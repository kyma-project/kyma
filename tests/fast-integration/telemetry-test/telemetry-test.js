const k8s = require('@kubernetes/client-node');
const {assert, expect} = require('chai');
const fs = require('fs');
const path = require('path');
const {
  k8sCoreV1Api,
  k8sApply,
  k8sDelete,
  waitForK8sObject,
} = require('../utils');
const {logsPresentInLoki} = require('../logging');
const {
  exposeGrafana,
  unexposeGrafana,
} = require('../monitoring');
const telemetryNamespace = 'kyma-system';
const defaultNamespace = 'default';
const testStartTimestamp = new Date().toISOString();
const invalidLogPipelineCR = loadResourceFromFile('./resources/pipelines/invalid-log-pipeline.yaml');
const parserLogPipelineCR = loadResourceFromFile('./resources/pipelines/valid-parser-log-pipeline.yaml');
const fooBarDeployment = loadResourceFromFile('./resources/deployments/regex_filter_deployment.yaml');

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

// async function prepareEnvironment() {
//   const logPipelinePromise = k8sApply(parserLogPipelineCR, telemetryNamespace);
//   const deploymentPromise = k8sApply(fooBarDeployment, defaultNamespace);
//   await logPipelinePromise;
//   await deploymentPromise;
// }

// async function cleanEnvironment() {
//   const logPipelinePromise = k8sDelete(parserLogPipelineCR, telemetryNamespace);
//   const deploymentPromise = k8sDelete(fooBarDeployment, defaultNamespace);
//   await logPipelinePromise;
//   await deploymentPromise;
// }

describe('Telemetry Operator tests', function() {
  before('Expose Grafana', async () => {
    // await prepareEnvironment();
    await exposeGrafana();
  });

  after('Unexpose Grafana', async () => {
    // await cleanEnvironment();
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

  it('Should push the logs to the loki output', async () => {
    const labels = '{job="telemetry-fluent-bit", namespace="kyma-system"}';
    const logsPresent = await logsPresentInLoki(labels, testStartTimestamp);
    assert.isTrue(logsPresent, 'No logs present in Loki');
  });

  it('Should parse the logs using regex', async () => {
    try {
      k8sApply(parserLogPipelineCR, telemetryNamespace);
      k8sApply(fooBarDeployment, defaultNamespace);
      const labels = '{job="telemetry-fluent-bit", namespace="default"}|json|pass="bar"|user="foo"';
      const logsPresent = await logsPresentInLoki(labels, testStartTimestamp);
      assert.isTrue(logsPresent, 'No parsed logs present in Loki');
    } catch (e) {
      assert.fail(e);
    }
  });
});


