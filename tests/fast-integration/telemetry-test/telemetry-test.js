const k8s = require('@kubernetes/client-node');
const {assert, expect} = require('chai');
const fs = require('fs');
const path = require('path');
const {
  k8sCoreV1Api,
  k8sApply,
  k8sDelete,
} = require('../utils');

const {checkLokiLogs, lokiPortForward} = require('../logging');
const telemetryNamespace = 'kyma-system';
const testStartTimestamp = new Date().toISOString();


function loadResourceFromFile(file) {
  const yaml = fs.readFileSync(path.join(__dirname, file), {
    encoding: 'utf8',
  });
  return k8s.loadAllYaml(yaml);
}

const invalidLogPipelineCR = loadResourceFromFile('./invalid-log-pipeline.yaml');
const logPipelineCR = loadResourceFromFile('./log-pipeline.yaml');

describe('Telemetry Operator tests', function() {
  let cancelPortForward;

  before(async function() {
    cancelPortForward = lokiPortForward();
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

  it('Should reject the invalid LogPipeline', async () => {
    try {
      await k8sApply(invalidLogPipelineCR, telemetryNamespace);
    } catch (e) {
      assert.equal(e.statusCode, 403);
      expect(e.body.message).to.have.string('denied the request', 'Invalid indentation level');
    };
  });

  it('should push the logs to the loki output', async () => {
    const labels = '{job="telemetry-fluent-bit"}';
    await checkLokiLogs(testStartTimestamp, labels);
  });

  after(async function() {
    cancelPortForward();
    await k8sDelete(logPipelineCR, telemetryNamespace);
  });
});
