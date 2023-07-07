const k8s = require('@kubernetes/client-node');
const fs = require('fs');
const path = require('path');
const {
  assert,
  expect,
} = require('chai');
const {
  getSecret,
  k8sCoreV1Api,
  k8sApply,
  k8sDelete,
  sleep,
  fromBase64,
  getGateway,
  getVirtualService,
  retryPromise,
  deployJaeger,
  deployLoki,
  waitForConfigMap,
} = require('../utils');
const {
  logsPresentInLoki,
  queryLoki,
} = require('../logging');
const {
  exposeGrafana,
  unexposeGrafana,
} = require('../monitoring');
const {
  loadTestData,
  waitForLogPipelineStatusRunning,
  waitForTracePipeline,
  waitForPodWithLabel,
  waitForTracePipelineStatusRunning,
} = require('./helpers');
const axios = require('axios');
const {getJaegerTracesForService, getJaegerServices} = require('../tracing/client');

async function getTracingTestAppUrl() {
  const vs = await getVirtualService('tracing-test', 'tracing-test-app');
  const host = vs.spec.hosts[0];
  return `https://${host}`;
}

async function callTracingTestApp() {
  const testAppUrl = await getTracingTestAppUrl();

  return retryPromise(async () => {
    return await axios.get(testAppUrl, {timeout: 10000});
  }, 5, 30);
}

async function prepareEnvironment() {
  async function k8sApplyFile(name, namespace) {
    await k8sApply(loadTestData(name), namespace);
  }

  await k8sApplyFile('http-backend-namespaces.yaml');
  await k8sApplyFile('http-backend.yaml', 'http-backend-1');
  await k8sApplyFile('http-backend.yaml', 'http-backend-2');
  await k8sApplyFile('regex-filter-deployment.yaml', 'default');
  await k8sApplyFile('logs-workload.yaml', 'default');
  await k8sApplyFile('logs-workload.yaml', 'kyma-system');
  await k8sApplyFile('secret-trace-endpoint.yaml', 'default');
  const jaegerYaml = fs.readFileSync(
      path.join(__dirname, '../test/fixtures/jaeger/jaeger.yaml'),
      {
        encoding: 'utf8',
      },
  );
  await deployJaeger(k8s.loadAllYaml(jaegerYaml));

  const lokiYaml = fs.readFileSync(
      path.join(__dirname, '../test/fixtures/loki/loki.yaml'),
      {
        encoding: 'utf-8',
      },
  );

  await deployLoki(k8s.loadAllYaml(lokiYaml));
}

async function cleanEnvironment() {
  async function k8sDeleteFile(name, namespace) {
    await k8sDelete(loadTestData(name), namespace);
  }

  await k8sDeleteFile('http-backend.yaml', 'http-backend-1');
  await k8sDeleteFile('http-backend.yaml', 'http-backend-2');
  await k8sDeleteFile('http-backend-namespaces.yaml');
  await k8sDeleteFile('regex-filter-deployment.yaml', 'default');
  await k8sDeleteFile('logs-workload.yaml', 'default');
  await k8sDeleteFile('logs-workload.yaml', 'kyma-system');
  await k8sDeleteFile('secret-trace-endpoint.yaml', 'default');
}

describe('Telemetry Operator', function() {
  const testStartTimestamp = new Date().toISOString();
  const defaultRetryDelayMs = 1000;
  const defaultRetries = 5;
  before('Prepare environment, expose Grafana', async function() {
    await prepareEnvironment();
    await exposeGrafana();
  });

  after('Clean environment, unexpose Grafana', async function() {
    await cleanEnvironment();
    await unexposeGrafana();
  });

  it('Should be ready', async function() {
    const podRes = await k8sCoreV1Api.listNamespacedPod(
        'kyma-system',
        'true',
        undefined,
        undefined,
        undefined,
        'control-plane=telemetry-operator',
    );
    const podList = podRes.body.items;
    assert.equal(podList.length, 1);

    const epRes = await k8sCoreV1Api.listNamespacedEndpoints(
        'kyma-system',
        'true',
        undefined,
        undefined,
        undefined,
        'control-plane=telemetry-operator',
    );
    const epList = epRes.body.items;
    assert.equal(epList.length, 2);
    assert.isNotEmpty(epList[0].subsets);
    assert.isNotEmpty(epList[0].subsets[0].addresses);
  });

  context('Configurable Logging', function() {
    context('Custom Loki LogPipeline', function() {
      it('Should be \'Running\'', async function() {
        await waitForLogPipelineStatusRunning('loki-test');
      });

      it('Should push system logs to Loki', async function() {
        const labels = '{namespace="kyma-system", job="telemetry-fluent-bit"}';
        const logsPresent = await logsPresentInLoki(labels, testStartTimestamp, 10);
        assert.isTrue(logsPresent, 'No logs present in Loki with namespace="kyma-system"');
      });
    });

    context('Webhook', function() {
      it('Should reject LogPipeline with unknown custom filter', async function() {
        const pipeline = loadTestData('logpipeline-custom-filter-unknown.yaml');

        try {
          await retryWithDelayForErrorCode((r) => k8sApply(pipeline), defaultRetryDelayMs, defaultRetries, 403);
          await k8sDelete(pipeline);
          assert.fail('Should not be able to apply a LogPipeline with an unknown custom filter');
        } catch (e) {
          assert.equal(e.statusCode, 403);
          expect(e.body.message).to.have.string('denied the request');
          const errMsg = 'section \'abc\' tried to instance a plugin name that don\'t exists';
          expect(e.body.message).to.have.string(errMsg);
        }
      });

      it('Should reject LogPipeline with denied custom filter', async function() {
        const pipeline = loadTestData('logpipeline-custom-filter-denied.yaml');

        try {
          await retryWithDelayForErrorCode((r) => k8sApply(pipeline), defaultRetryDelayMs, defaultRetries, 403);
          await k8sDelete(pipeline);
          assert.fail('Should not be able to apply a LogPipeline with a denied custom filter');
        } catch (e) {
          assert.equal(e.statusCode, 403);
          expect(e.body.message).to.have.string('denied the request');
          const errMsg = 'plugin \'kubernetes\' is forbidden';
          expect(e.body.message).to.have.string(errMsg);
        }
      });
    });

    context('LogParser', function() {
      const parser = loadTestData('logparser-regex.yaml');
      const parserName = parser[0].metadata.name;

      it(`Should create LogParser '${parserName}'`, async function() {
        await retryWithDelay( (r)=> k8sApply(parser), defaultRetryDelayMs, defaultRetries);
      });

      it('Should parse the logs using regex', async function() {
        try {
          const labels = '{namespace="default"}|json|pass="bar"|user="foo"';
          const logsPresent = await logsPresentInLoki(labels, testStartTimestamp);
          assert.isTrue(logsPresent, 'No parsed logs present in Loki');
        } catch (e) {
          assert.fail(e);
        }
      });

      it(`Should delete LogParser '${parserName}'`, async function() {
        await k8sDelete(parser);
      });
    });

    context('LogPipeline', function() {
      context('HTTP Output', function() {
        const backend1Secret = loadTestData('http-backend-1-secret.yaml');
        const backend1Host = backend1Secret[0].stringData.host;
        const backend2Secret = loadTestData('http-backend-2-secret.yaml');
        const backend2Host = backend2Secret[0].stringData.host;

        it(`Should create host secret with host set to '${backend1Host}'`, async function() {
          await k8sApply(loadTestData('http-backend-1-secret.yaml'));
        });

        const pipeline = loadTestData('logpipeline-output-http.yaml');
        const pipelineName = pipeline[0].metadata.name;

        it(`Should create LogPipeline '${pipelineName}'`, async function() {
          await k8sApply(pipeline);
          await waitForLogPipelineStatusRunning(pipelineName);
        });

        it(`Should push logs to '${backend1Host}'`, async function() {
          const labels = '{namespace="http-backend-1"}';
          const logsPresent = await logsPresentInLoki(labels, testStartTimestamp);
          assert.isTrue(logsPresent, 'No logs received by mockserver present in Loki');
        });

        it(`Should update host secret with host set to '${backend2Host}'`, async function() {
          await k8sApply(loadTestData('http-backend-2-secret.yaml'));
          await sleep(5000);
          await waitForLogPipelineStatusRunning(pipelineName);
        });

        it(`Should detect secret update and push logs to '${backend2Host}'`, async function() {
          const labels = '{namespace="http-backend-2"}';
          const logsPresent = await logsPresentInLoki(labels, testStartTimestamp);
          assert.isTrue(logsPresent, 'No logs received by mockserver present in Loki');
        });

        it(`Should delete LogPipeline '${pipelineName}'`, async function() {
          await k8sDelete(pipeline);
        });
      });

      context('Custom Output', function() {
        const backend1Secret = loadTestData('http-backend-1-secret.yaml');
        const backend1Host = backend1Secret[0].stringData.host;
        const backend2Secret = loadTestData('http-backend-2-secret.yaml');
        const backend2Host = backend2Secret[0].stringData.host;

        it(`Should create host secret with host set to '${backend1Host}'`, async function() {
          await k8sApply(loadTestData('http-backend-1-secret.yaml'));
        });

        const pipeline = loadTestData('logpipeline-output-custom.yaml');
        const pipelineName = pipeline[0].metadata.name;

        it(`Should create LogPipeline '${pipelineName}'`, async function() {
          await retryWithDelay( (r) => k8sApply(pipeline), defaultRetryDelayMs, defaultRetries);
          await waitForLogPipelineStatusRunning(pipelineName);
        });

        it(`Should push logs to '${backend1Host}'`, async function() {
          const labels = '{namespace="http-backend-1"}';
          const logsPresent = await logsPresentInLoki(labels, testStartTimestamp);
          assert.isTrue(logsPresent, 'No logs received by mockserver present in Loki');
        });

        it(`Should update host secret with host set to '${backend2Host}'`, async function() {
          await k8sApply(loadTestData('http-backend-2-secret.yaml'));
          await sleep(5000);
          await waitForLogPipelineStatusRunning(pipelineName);
        });

        it(`Should detect secret update and push logs to '${backend2Host}'`, async function() {
          const labels = '{namespace="http-backend-2"}';
          const logsPresent = await logsPresentInLoki(labels, testStartTimestamp);
          assert.isTrue(logsPresent, 'No logs received by mockserver present in Loki');
        });

        it(`Should delete LogPipeline '${pipelineName}'`, async function() {
          await k8sDelete(pipeline);
        });
      });

      context('Input', function() {
        context('Drop annotations, keep labels', function() {
          const pipeline = loadTestData('logpipeline-input-keep-labels.yaml');
          const pipelineName = pipeline[0].metadata.name;

          it(`Should create LogPipeline '${pipelineName}'`, async function() {
            await k8sApply(pipeline);
            await waitForLogPipelineStatusRunning(pipelineName);
          });

          it(`Should push only labels to Loki`, async function() {
            const labels = '{job="drop-annotations-keep-labels-telemetry-fluent-bit", container="flog"}';
            const found = await logsPresentInLoki(labels, testStartTimestamp);
            assert.isTrue(found, `No logs in Loki with labels: ${labels}`);

            const responseBody = await queryLoki(labels, testStartTimestamp);
            const entry = JSON.parse(responseBody.data.result[0].values[0][1]);
            assert.hasAnyKeys(entry, 'kubernetes', `No kubernetes metadata in ${entry}`);
            const k8smeta = entry['kubernetes'];
            assert.doesNotHaveAnyKeys(k8smeta, 'annotations', `Annotations found in ${JSON.stringify(k8smeta)}`);
            assert.hasAnyKeys(k8smeta, 'labels', `No labels in ${JSON.stringify(k8smeta)}`);
          });

          it(`Should delete LogPipeline '${pipelineName}'`, async function() {
            await k8sDelete(pipeline);
          });
        });

        context('Keep annotations, drop labels', function() {
          const pipeline = loadTestData('logpipeline-input-drop-labels.yaml');
          const pipelineName = pipeline[0].metadata.name;

          it(`Should create LogPipeline '${pipelineName}'`, async function() {
            await retryWithDelay( (r) => k8sApply(pipeline), defaultRetryDelayMs, defaultRetries);
            await waitForLogPipelineStatusRunning(pipelineName);
          });

          it(`Should push only annotations to Loki`, async function() {
            const labels = '{job="keep-annotations-drop-labels-telemetry-fluent-bit", container="flog"}';
            const found = await logsPresentInLoki(labels, testStartTimestamp);
            assert.isTrue(found, `No logs in Loki with labels: ${labels}`);

            const responseBody = await queryLoki(labels, testStartTimestamp);
            const entry = JSON.parse(responseBody.data.result[0].values[0][1]);
            assert.hasAnyKeys(entry, 'kubernetes', `No kubernetes metadata in ${entry}`);
            const k8smeta = entry['kubernetes'];
            assert.doesNotHaveAnyKeys(k8smeta, 'labels', `Labels found in ${JSON.stringify(k8smeta)}`);
            assert.hasAnyKeys(k8smeta, 'annotations', `No annotations in ${JSON.stringify(k8smeta)}`);
          });

          it(`Should delete LogPipeline '${pipelineName}'`, async function() {
            await k8sDelete(pipeline);
          });
        });

        context('Containers Exclude', function() {
          const pipeline = loadTestData('logpipeline-input-containers-exclude.yaml');
          const pipelineName = pipeline[0].metadata.name;

          it(`Should create LogPipeline '${pipelineName}'`, async function() {
            await k8sApply(pipeline);
            await waitForLogPipelineStatusRunning(pipelineName);
          });

          it(`Should not push any system logs to Loki`, async function() {
            const labels = '{namespace="kyma-system", job="exclude-istio-proxy-telemetry-fluent-bit"}';
            const logsFound = await logsPresentInLoki(labels, testStartTimestamp, 3);
            assert.isFalse(logsFound, `No logs must present in Loki for labels: ${labels}`);
          });

          it(`Should not push any istio-proxy logs to Loki`, async function() {
            const labels = '{container="istio-proxy", job="exclude-istio-proxy-telemetry-fluent-bit"}';
            const logsFound = await logsPresentInLoki(labels, testStartTimestamp, 3);
            assert.isFalse(logsFound, `No logs must present in Loki for labels: ${labels}`);
          });

          it(`Should delete LogPipeline '${pipelineName}'`, async function() {
            await k8sDelete(pipeline);
          });
        });
      });
    });
  });

  context('Configurable Tracing', function() {
    context('Configurable Tracing', function() {
      context('TracePipeline', function() {
        const jaeger = loadTestData('tracepipeline-jaeger.yaml');
        const firstPipeline = loadTestData('tracepipeline-output-otlp-secret-ref-1.yaml');
        const firstPipelineName = firstPipeline[0].metadata.name;

        it(`Should clean up TracePipeline jaeger`, async function() {
          await k8sDelete(jaeger);
        });

        it(`Should create TracePipeline '${firstPipelineName}'`, async function() {
          await k8sApply(firstPipeline);
          await waitForTracePipeline(firstPipelineName);
        });

        it('Should be \'Running\'', async function() {
          await waitForTracePipelineStatusRunning(firstPipelineName);
        });

        it('Should have ready trace collector pods', async () => {
          await waitForPodWithLabel('app.kubernetes.io/name', 'telemetry-trace-collector', 'kyma-system');
        });

        it('Should have created telemetry-trace-collector secret', async () => {
          const secret = await getSecret('telemetry-trace-collector', 'kyma-system');
          assert.equal(secret.data.OTLP_ENDPOINT_OTLP_OUTPUT_ENDPOINT_SECRET_REF_1, 'aHR0cDovL25vLWVuZHBvaW50');
        });

        it(`Should reflect secret ref change in telemetry-trace-collector secret and pod restart`, async function() {
          const podRes = await k8sCoreV1Api.listNamespacedPod(
              'kyma-system',
              'true',
              undefined,
              undefined,
              undefined,
              'app.kubernetes.io/name=telemetry-trace-collector',
          );
          const podList = podRes.body.items;

          await k8sApply(loadTestData('secret-patched-trace-endpoint.yaml'), 'default');
          await sleep(5*1000);
          const secret = await getSecret('telemetry-trace-collector', 'kyma-system');
          assert.equal(secret.data.OTLP_ENDPOINT_OTLP_OUTPUT_ENDPOINT_SECRET_REF_1, 'aHR0cDovL2Fub3RoZXItZW5kcG9pbnQ=');

          const newPodRes = await k8sCoreV1Api.listNamespacedPod(
              'kyma-system',
              'true',
              undefined,
              undefined,
              undefined,
              'app.kubernetes.io/name=telemetry-trace-collector',
          );
          const newPodList = newPodRes.body.items;
          assert.notDeepEqual(
              newPodList,
              podList,
              'telemetry-trace-collector has not been  restarted after Secret change',
          );
        });

        it(`Should delete first TracePipeline '${firstPipeline}'`, async function() {
          await k8sDelete(firstPipeline);
        });
      });

      context('Debuggability', function() {
        const overrideConfig = loadTestData('override-config.yaml');
        const pipeline = loadTestData('tracepipeline-output-otlp.yaml');
        const pipelineName = pipeline[0].metadata.name;
        it(`Creates a tracepipeline`, async function() {
          await k8sApply(pipeline);
          await waitForTracePipeline(pipelineName);
          await waitForTracePipelineStatusRunning(pipelineName);
        });

        it('Should have created telemetry-trace-collector secret', async () => {
          const secret = await getSecret('telemetry-trace-collector', 'kyma-system');
          assert.equal(fromBase64(secret.data.OTLP_ENDPOINT_TEST_TRACE), 'http://foo-bar');
        });

        it(`Should create override configmap with paused flag`, async function() {
          await retryWithDelay( (r) => k8sApply(overrideConfig), defaultRetryDelayMs, defaultRetries);
          await waitForConfigMap('telemetry-override-config', 'kyma-system');
        });

        it(`Tries to change the otlp endpoint`, async function() {
          await sleep(5*1000);
          pipeline[0].spec.output.otlp.endpoint.value = 'http://another-foo';
          await retryWithDelay( (r) => k8sApply(pipeline), defaultRetryDelayMs, defaultRetries);
        });

        it(`Should not change the OTLP endpoint in the telemetry-trace-collector secret in paused state`, async () => {
          await sleep(5*1000);
          const secret = await getSecret('telemetry-trace-collector', 'kyma-system');
          assert.equal(fromBase64(secret.data.OTLP_ENDPOINT_TEST_TRACE), 'http://foo-bar');
        });

        it(`Deletes the override configmap`, async function() {
          await k8sDelete(overrideConfig);
        });

        it(`Tries to change the otlp endpoint again`, async function() {
          await sleep(10*1000);
          pipeline[0].spec.output.otlp.endpoint.value = 'http://another-foo-bar';
          await k8sApply(pipeline);
          await waitForTracePipeline(pipelineName);
          await waitForTracePipelineStatusRunning(pipelineName);
        });

        it(`Should now change the OTLP endpoint in the telemetry-trace-collector secret`, async function() {
          await sleep(5*1000);
          const secret = await getSecret('telemetry-trace-collector', 'kyma-system');
          assert.equal(fromBase64(secret.data.OTLP_ENDPOINT_TEST_TRACE), 'http://another-foo-bar');
        });

        it(`Should delete TracePipeline`, async function() {
          await k8sDelete(pipeline);
        });
      });

      context('Filter Processor', function() {
        const testApp = loadTestData('tracepipeline-test-app.yaml');
        const testAppIstioPatch = loadTestData('tracepipeline-test-istio-telemetry-patch.yaml');

        it(`Should create test app`, async function() {
          const kymaGateway = await getGateway('kyma-system', 'kyma-gateway');
          let kymaHostUrl = kymaGateway.spec.servers[0].hosts[0];
          kymaHostUrl = kymaHostUrl.replace('*', 'tracing-test-app');
          for (const resource of testApp ) {
            if (resource.kind == 'VirtualService') {
              resource.spec.hosts[0] = kymaHostUrl;
            }
          }
          await retryWithDelay( (r) => k8sApply(testApp), defaultRetryDelayMs, defaultRetries);
          await retryWithDelay( (r) =>k8sApply(testAppIstioPatch), defaultRetryDelayMs, defaultRetries);
          await waitForPodWithLabel('app', 'tracing-test-app', 'tracing-test');
        });

        it(`Should call test app and produce spans`, async function() {
          for (let i=0; i < 10; i++) {
            await retryWithDelay(callTracingTestApp, defaultRetryDelayMs, defaultRetries);
            await sleep(500);
          }
        });

        it(`Should filter out noisy spans`, async function() {
          const services = await retryWithDelay(async function() {
            const services = await getJaegerServices();
            if (services.data.length > 0) {
              return services;
            }

            throw services;
          }, defaultRetryDelayMs, defaultRetries);
          assert.isFalse(services.data.includes('grafana.kyma-system'), 'spans are present for grafana');
          assert.isFalse(services.data.includes('telemetry-fluent-bit.kyma-system'),
              'spans are present for fluent-bit');
          assert.isFalse(services.data.includes('loki.kyma-system'), 'spans are present for loki');
        });

        it(`Should find test spans`, async function() {
          const testAppTraces = await retryWithDelay( async (r) => {
            const testAppTraces = await getJaegerTracesForService('tracing-test-app', 'tracing-test');
            if (testAppTraces.data.length > 0) {
              return testAppTraces;
            }

            throw testAppTraces;
          }, defaultRetryDelayMs, 20);
          assert.isTrue(testAppTraces.data.length > 0, 'No spans present for test application "tracing-test-app"');
        });

        it(`Should delete test setup`, async function() {
          testAppIstioPatch[0].spec.tracing[0].randomSamplingPercentage = 1;
          await k8sApply(testAppIstioPatch);
          await k8sDelete(testApp);
        });
      });
    });
  });
});

const wait = (ms) => new Promise((resolve) => {
  setTimeout(() => resolve(), ms);
});

const retryWithDelay = (operation, delay, retries) => new Promise((resolve, reject) => {
  return operation()
      .then(resolve)
      .catch((reason) => {
        if (retries > 0) {
          return wait(delay)
              .then(retryWithDelay.bind(null, operation, delay, retries - 1))
              .then(resolve)
              .catch(reject);
        }
        return reject(reason);
      });
});

const retryWithDelayForErrorCode = (operation, delay, retries, expectedErrorCode) => new Promise((resolve, reject) => {
  return operation()
      .then(resolve)
      .catch((reason) => {
        if (reason.statusCode !== undefined && reason.statusCode === expectedErrorCode) {
          return reject(reason);
        }
        if (retries > 0) {
          return wait(delay)
              .then(retryWithDelay.bind(null, operation, delay, retries - 1, expectedErrorCode))
              .then(resolve)
              .catch(reject);
        }
        return reject(reason);
      });
});
