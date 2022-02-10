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
} = require('../utils');
const {mockServerClient} = require('mockserver-client');

const mockServerPort = 1080;
const telemetryNamespace = 'kyma-system';
const mockNamespace = 'mockserver';
const fluentBitName = 'telemetry-fluent-bit';
const logPipelineCR = (() => {
  const yaml = fs.readFileSync(path.join(__dirname, './log-pipeline.yaml'), {
    encoding: 'utf8',
  });
  return k8s.loadAllYaml(yaml);
})();
const mockserverResources = (() => {
  const yaml = fs.readFileSync(path.join(__dirname, './mockserver/resources.yaml'), {
    encoding: 'utf8',
  });
  return k8s.loadAllYaml(yaml);
})();

function checkMockserverWasCalled(wasCalled) {
  const args = wasCalled ? [1] : [0, 0];
  const not = wasCalled ? 'not ' : '';
  return mockServerClient('localhost', mockServerPort)
      .verify(
          {
            path: '/',
          },
          ...args,
      )
      .then(
          function() {},
          function(error) {
            console.log(error);
            assert.fail(`"The HTTP endpoint was ${not}called`);
          },
      );
}

describe('Telemetry Operator tests', function() {
  let cancelPortForward;

  before(async function() {
    console.log('Creating namespace');
    await k8sApply([namespaceObj(mockNamespace)]);
    console.log('Waiting for namespace');
    await waitForNamespace(mockNamespace);
    console.log('Apply mockserver resources');
    await k8sApply(mockserverResources, mockNamespace);
    console.log('Waiting for mockserver resources');
    await waitForDeployment('mockserver', mockNamespace);
    console.log('List all mockserver resources');
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

  // it('Should not receive HTTP traffic', () => {
  //   return checkMockserverWasCalled(false);
  // });

  it('Apply HTTP output plugin to fluent-bit', async () => {
    await k8sApply(logPipelineCR, telemetryNamespace);
    await waitForDaemonSet(fluentBitName, telemetryNamespace, 10000);
  });

  it('Should receive HTTP traffic from fluent-bit', () => {
    return checkMockserverWasCalled(true);
  });

  after(async function() {
    cancelPortForward();
    await k8sDelete(logPipelineCR, telemetryNamespace);
    // await k8sCoreV1Api.deleteNamespace(mockNamespace);
  });
});
