const k8s = require('@kubernetes/client-node');
const {assert, expect} = require('chai');
const fs = require('fs');
const path = require('path');
const {
  k8sCoreV1Api,
  k8sApply,
  k8sDelete,
  kubectlPortForward,
  namespaceObj,
  waitForK8sObject,
  waitForDeployment,
  waitForNamespace,
  sleep,
} = require('../utils');
const {mockServerClient} = require('mockserver-client');

const mockServerPort = 1080;
const telemetryNamespace = 'kyma-system';
const mockNamespace = 'mockserver';

function loadResourceFromFile(file) {
  const yaml = fs.readFileSync(path.join(__dirname, file), {
    encoding: 'utf8',
  });
  return k8s.loadAllYaml(yaml);
}

function checkLastCondition(logPipeline, conditionType) {
  const conditions = logPipeline.status.conditions;
  if (conditions.length == 0) {
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
          watchObj.metadata.name == name && checkLastCondition(watchObj, lastConditionType)
        );
      },
      timeout,
      `Waiting for log pipeline ${name} timeout (${timeout} ms)`,
  );
}

const logPipelineCR = loadResourceFromFile('./log-pipeline.yaml');
const invalidLogPipelineCR = loadResourceFromFile('./invalid-log-pipeline.yaml');
const mockserverResources = loadResourceFromFile('./mockserver-resources.yaml');

function assertMockserverWasCalled() {
  return mockServerClient('localhost', mockServerPort)
      .verify(
          {
            path: '/',
          },
          1,
      )
      .then(
          function() {},
          function(error) {
            console.log(error);
            assert.fail('The HTTP endpoint was not called');
          },
      );
}

describe('Telemetry Operator tests', function() {
  let cancelPortForward;

  before(async function() {
    await k8sApply([namespaceObj(mockNamespace)]);
    await waitForNamespace(mockNamespace);
    await k8sApply(mockserverResources, mockNamespace);
    await waitForDeployment('mockserver', mockNamespace);
    const {body} = await k8sCoreV1Api.listNamespacedPod(mockNamespace);
    const mockPod = body.items[0].metadata.name;
    cancelPortForward = kubectlPortForward(
        mockNamespace,
        mockPod,
        mockServerPort,
    );
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

  it('Should create valid LogPipeline with HTTP output plugin', async () => {
    await k8sApply(logPipelineCR, telemetryNamespace);
    await waitForLogPipelineStatusCondition('logpipeline-test', 'Pending', 20000);
    await waitForLogPipelineStatusCondition('logpipeline-test', 'Running', 180000);
  });

  it('Mockserver should receive HTTP traffic from fluent-bit', async () => {
    await sleep(30000);
    await assertMockserverWasCalled(true);
  });

  after(async function() {
    cancelPortForward();
    await k8sDelete(logPipelineCR, telemetryNamespace);
    await k8sCoreV1Api.deleteNamespace(mockNamespace);
  });
});
