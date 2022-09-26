const {
  assert,
  expect,
} = require('chai');
const {
  k8sCoreV1Api,
  k8sApply,
  k8sDelete,
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
} = require('./helpers');

async function prepareEnvironment() {
  async function k8sApplyFile(name, namespace) {
    await k8sApply(loadTestData(name), namespace);
  }

  await k8sApplyFile('mockserver.yaml', 'mockserver');
  await k8sApplyFile('regex-filter-deployment.yaml', 'default');
  await k8sApplyFile('logs-workload.yaml', 'default');
  await k8sApplyFile('logs-workload.yaml', 'kyma-system');
}

async function cleanEnvironment() {
  async function k8sDeleteFile(name, namespace) {
    await k8sDelete(loadTestData(name), namespace);
  }

  await k8sDeleteFile('mockserver.yaml', 'mockserver');
  await k8sDeleteFile('regex-filter-deployment.yaml', 'default');
  await k8sDeleteFile('logs-workload.yaml', 'default');
  await k8sDeleteFile('logs-workload.yaml', 'kyma-system');
}

describe('Telemetry Operator', function() {
  const testStartTimestamp = new Date().toISOString();

  before('Prepare environment, expose Grafana', async function() {
    await prepareEnvironment();
    await exposeGrafana();
  });

  after('Clean environment, unexpose Grafana', async function() {
    await cleanEnvironment();
    await unexposeGrafana();
  });

  it('Should be ready', async function() {
    const res = await k8sCoreV1Api.listNamespacedPod(
        'kyma-system',
        'true',
        undefined,
        undefined,
        undefined,
        'control-plane=telemetry-operator',
    );
    const podList = res.body.items;
    assert.equal(podList.length, 1);
  });

  context('Configurable Logging', function() {
    context('Default Loki LogPipeline', function() {
      it('Should be \'Running\'', async function() {
        await waitForLogPipelineStatusRunning('loki');
      });

      it('Should push system logs to Kyma Loki', async function() {
        const labels = '{namespace="kyma-system", job="telemetry-fluent-bit"}';
        const logsPresent = await logsPresentInLoki(labels, testStartTimestamp, 3);
        assert.isTrue(logsPresent, 'No logs present in Loki with namespace="kyma-system"');
      });
    });

    context('Webhook', function() {
      it('Should reject LogPipeline with unknown custom filter', async function() {
        const pipeline = loadTestData('logpipeline-custom-filter-unknown.yaml');

        try {
          await k8sApply(pipeline);
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
          await k8sApply(pipeline);
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
        await k8sApply(parser);
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
        const pipeline = loadTestData('logpipeline-output-http.yaml');
        const pipelineName = pipeline[0].metadata.name;

        it(`Should create LogPipeline '${pipelineName}'`, async function() {
          await k8sApply(pipeline);
          await waitForLogPipelineStatusRunning(pipelineName);
        });

        it('Should push logs to the HTTP mockserver', async function() {
          const labels = '{namespace="mockserver"}';
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
            const labels = '{job="drop-annotations-keep-labels-telemetry-fluent-bit"}';
            const found = await logsPresentInLoki(labels, testStartTimestamp);
            assert.isTrue(found, `No logs present in Loki for labels: ${labels}`);

            const responseBody = await queryLoki(labels, testStartTimestamp);
            const entry = JSON.parse(responseBody.data.result[0].values[0][1]);
            assert.isTrue('kubernetes' in entry, `No kubernetes metadata present in log entry: ${entry} `);
            expect(entry['kubernetes']).not.to.have.property('annotations');
            expect(entry['kubernetes']).to.have.property('labels');
          });

          it(`Should delete LogPipeline '${pipelineName}'`, async function() {
            await k8sDelete(pipeline);
          });
        });

        context('Keep annotations, drop labels', function() {
          const pipeline = loadTestData('logpipeline-input-drop-labels.yaml');
          const pipelineName = pipeline[0].metadata.name;

          it(`Should create LogPipeline '${pipelineName}'`, async function() {
            await k8sApply(pipeline);
            await waitForLogPipelineStatusRunning(pipelineName);
          });

          it(`Should push only annotations to Loki`, async function() {
            const labels = '{job="keep-annotations-drop-labels-telemetry-fluent-bit"}';
            const found = await logsPresentInLoki(labels, testStartTimestamp);
            assert.isTrue(found, `No logs present in Loki for labels: ${labels}`);

            const responseBody = await queryLoki(labels, testStartTimestamp);
            assert.isTrue(responseBody.data.result.length > 0, `No logs present in Loki for labels: ${labels}`);

            const entry = JSON.parse(responseBody.data.result[0].values[0][1]);
            assert.isTrue('kubernetes' in entry, `No kubernetes metadata present in log entry: ${entry} `);
            expect(entry['kubernetes']).not.to.have.property('labels');
            expect(entry['kubernetes']).to.have.property('annotations');
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
});
