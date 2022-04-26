const k8s = require('@kubernetes/client-node');
const {assert, expect} = require('chai');
const fs = require('fs');
const path = require('path');
const {
  k8sCoreV1Api,
  k8sApply,
  waitForK8sObject,
} = require('../utils');
const {logsPresentInLoki} = require('../logging');
const {
  exposeGrafana,
  unexposeGrafana,
} = require('../monitoring');
const telemetryNamespace = 'kyma-system';
const testStartTimestamp = new Date().toISOString();
const invalidLogPipelineCR = loadResourceFromFile('./invalid-log-pipeline.yaml');

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


describe('Telemetry Operator tests', function() {
  before('Prepare Grafana', async () => {
    await exposeGrafana();
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
    } catch (e) {
      assert.equal(e.statusCode, 403);
      expect(e.body.message).to.have.string('denied the request', 'Invalid indentation level');
    }
  });

  it('Should push the logs to the loki output', async () => {
    const labels = '{job="telemetry-fluent-bit"}';
    const logsPresent = await logsPresentInLoki(labels, testStartTimestamp);
    assert.isTrue(logsPresent, 'No logs present in Loki');
  });

  after('Cleanup Grafana', async () => {
    await unexposeGrafana();
  });
});
