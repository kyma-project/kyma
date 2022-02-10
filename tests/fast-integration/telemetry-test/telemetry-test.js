const k8s = require('@kubernetes/client-node');
const {assert} = require('chai');
const fs = require('fs');
const path = require('path');
const {
  k8sCoreV1Api,
  k8sApply,
  k8sDelete,
  kubectlPortForward,
  namespaceObj,
  waitForDaemonSet,
  waitForDeployment,
  waitForNamespace,
  sleep,
} = require('../utils');
const {mockServerClient} = require('mockserver-client');

const mockServerPort = 1080;
const telemetryNamespace = 'kyma-system';
const mockNamespace = 'mockserver';
const fluentBitName = 'telemetry-fluent-bit';

function loadResourceFromFile(file) {
  const yaml = fs.readFileSync(path.join(__dirname, file), {
    encoding: 'utf8',
  });
  return k8s.loadAllYaml(yaml);
}

const logPipelineCR = loadResourceFromFile('./log-pipeline.yaml');
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
        'control-plane=telemetry-operator-controller-manager',
    );
    const podList = res.body.items;
    assert.equal(podList.length, 1);
  });

  it('Apply HTTP output plugin to fluent-bit', async () => {
    await k8sApply(logPipelineCR, telemetryNamespace);
    await waitForDaemonSet(fluentBitName, telemetryNamespace, 30000);
  });

  it('Should receive HTTP traffic from fluent-bit', async () => {
    await sleep(30000);
    await assertMockserverWasCalled(true);
  });

  after(async function() {
    cancelPortForward();
    await k8sDelete(logPipelineCR, telemetryNamespace);
    await k8sCoreV1Api.deleteNamespace(mockNamespace);
  });
});
