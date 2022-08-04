const k8s = require('@kubernetes/client-node');
const {assert, expect} = require('chai');
const fs = require('fs');
const path = require('path');
const {
  k8sCoreV1Api,
  k8sDynamicApi,
  k8sApply,
  k8sDelete,
  sleep,
  waitForK8sObject,
} = require('../utils');
const {logsPresentInLoki} = require('../logging');
const {
  exposeGrafana,
  unexposeGrafana,
} = require('../monitoring');

const telemetryNamespace = 'kyma-system';
const defaultNamespace = 'default';
const mockserverNamespace = 'mockserver';
const testStartTimestamp = new Date().toISOString();
const invalidLogPipelineCR = loadResourceFromFile('./resources/pipelines/invalid-log-pipeline.yaml');
const parserLogPipelineCR = loadResourceFromFile('./resources/pipelines/valid-parser-log-pipeline.yaml');
const regexFilterDeployment = loadResourceFromFile('./resources/deployments/regex_filter_deployment.yaml');
const mockserverDeployment = loadResourceFromFile('./resources/deployments/mockserver.yaml');
const httpLogPipelineCR = loadResourceFromFile('./resources/pipelines/http-log-pipeline.yaml');

function loadResourceFromFile(file) {
  const yaml = fs.readFileSync(path.join(__dirname, file), {
    encoding: 'utf8',
  });
  return k8s.loadAllYaml(yaml);
}

function checkLastCondition(logPipeline, conditionType) {
  const conditions = logPipeline.status.conditions;
  if (conditions.length === 0) {
    return false;
  }
  const lastCondition = conditions[conditions.length - 1];
  return lastCondition.type === conditionType;
}

function waitForLogPipelineStatusCondition(name, lastConditionType, timeout) {
  return waitForK8sObject(
      '/apis/telemetry.kyma-project.io/v1alpha1/logpipelines',
      {},
      (_type, watchObj, _) => {
        return (
          watchObj.metadata.name === name && checkLastCondition(watchObj, lastConditionType)
        );
      },
      timeout,
      `Waiting for log pipeline ${name} timeout (${timeout} ms)`,
  );
}

async function getLogPipeline(name) {
  const path = `/apis/telemetry.kyma-project.io/v1alpha1/logpipelines/${name}`;
  const response = await k8sDynamicApi.requestPromise({
    url: k8sDynamicApi.basePath + path,
  });
  return JSON.parse(response.body);
}

async function updateLogPipeline(logPipeline) {
  const options = {
    headers: {'Content-type': 'application/merge-patch+json'},
  };

  await k8sDynamicApi.patch(
      logPipeline,
      undefined,
      undefined,
      undefined,
      undefined,
      options,
  );
}

async function prepareEnvironment() {
  const lokiLogPipelinePromise = k8sApply(parserLogPipelineCR, telemetryNamespace);
  const httpLogPipelinePromise = k8sApply(httpLogPipelineCR, telemetryNamespace);
  const mockserverPromise = k8sApply(mockserverDeployment, mockserverNamespace);
  const deploymentPromise = k8sApply(regexFilterDeployment, defaultNamespace);
  await lokiLogPipelinePromise;
  await httpLogPipelinePromise;
  await mockserverPromise;
  await deploymentPromise;
}

async function cleanEnvironment() {
  const logPipelinePromise = k8sDelete(parserLogPipelineCR, telemetryNamespace);
  const mockserverPromise = k8sDelete(mockserverDeployment, mockserverNamespace);
  const deploymentPromise = k8sDelete(regexFilterDeployment, defaultNamespace);
  const httpLogPipelinePromise = k8sDelete(httpLogPipelineCR, telemetryNamespace);
  await logPipelinePromise;
  await mockserverPromise;
  await httpLogPipelinePromise;
  await deploymentPromise;
}

describe('Telemetry Operator tests, prepare the environment', function() {
  before('Expose Grafana', async () => {
    await prepareEnvironment();
    await exposeGrafana();
  });

  after('Unexpose Grafana, clean the environment', async () => {
    await cleanEnvironment();
    await unexposeGrafana();
  });

  it('Operator should be ready', async () => {
    const res = await k8sCoreV1Api.listNamespacedPod(
        telemetryNamespace,
        'true',
        undefined,
        undefined,
        undefined,
        'control-plane=telemetry-operator',
    );
    const podList = res.body.items;
    assert.equal(podList.length, 1);
  });

  it('Loki LogPipeline should have Running condition', async () => {
    await waitForLogPipelineStatusCondition('loki', 'Running', 180000);
  });

  it('Should reject the invalid LogPipeline', async () => {
    try {
      await k8sApply(invalidLogPipelineCR, telemetryNamespace);
      await k8sDelete(invalidLogPipelineCR, telemetryNamespace);
      assert.fail('Should not be able to apply invalid LogPipeline');
    } catch (e) {
      assert.equal(e.statusCode, 403);
      expect(e.body.message).to.have.string('denied the request');
      const errMsg = 'section \'abc\' tried to instance a plugin name that don\'t exists';
      expect(e.body.message).to.have.string(errMsg);
    }
  });

  it('Should exclude system namespace by default', async () => {
    await sleep(5 * 1000);
    const labels = '{job="telemetry-fluent-bit", namespace="kyma-system"}';
    const logsPresent = await logsPresentInLoki(labels, testStartTimestamp);
    assert.isFalse(logsPresent, 'No logs present in Loki');
  });

  it('Should parse the logs using regex', async () => {
    try {
      const labels = '{job="telemetry-fluent-bit", namespace="default"}|json|pass="bar"|user="foo"';
      const logsPresent = await logsPresentInLoki(labels, testStartTimestamp);
      assert.isTrue(logsPresent, 'No parsed logs present in Loki');
    } catch (e) {
      assert.fail(e);
    }
  });

  it('HTTP LogPipeline should have Running condition', async () => {
    await waitForLogPipelineStatusCondition('http-mockserver', 'Running', 180000);
  });

  it('Should push the logs to the http mockserver', async () => {
    // The mockserver prints received logs to stdout, which should finally be pushed to Loki by the other pipeline
    const labels = '{job="telemetry-fluent-bit", namespace="mockserver"}';
    const logsPresent = await logsPresentInLoki(labels, testStartTimestamp);
    assert.isTrue(logsPresent, 'No logs received by mockserver present in Loki');
  });

  it('Include kyma-system namespace on loki pipeline ', async () => {
    const lokiPipeline = await getLogPipeline('loki');
    lokiPipeline.spec.input.application.includeSystemNamespaces = true;
    await updateLogPipeline(lokiPipeline);

    await sleep(10 * 1000);
    const labels = '{job="telemetry-fluent-bit", namespace="kyma-system"}';
    const logsPresent = await logsPresentInLoki(labels, testStartTimestamp);
    assert.isTrue(logsPresent, 'No kyma-system logs present in Loki');
  });
});
