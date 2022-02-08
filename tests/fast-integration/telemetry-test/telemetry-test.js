const k8s = require('@kubernetes/client-node');
const {assert} = require('chai');
const fs = require('fs');
const path = require('path');
const {
  waitForDaemonSet,
  waitForDeployment,
  k8sCoreV1Api,
  k8sApply,
  k8sDelete,
  kubectlPortForward,
} = require('../utils');
const mockServerClient = require('mockserver-client').mockServerClient;
const mockServerPort = 1080;

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

function loadCR(filepath) {
  const _logPipelineYaml = fs.readFileSync(path.join(__dirname, filepath), {
    encoding: 'utf8',
  });
  return k8s.loadAllYaml(_logPipelineYaml);
}

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
  const telemetryNamespace = 'kyma-system'; // operator flag 'fluent-bit-ns' is set to kyma-system
  const mockNamespace = 'mockserver';
  let cancelPortForward = null;
  const fluentBitName = 'telemetry-fluent-bit';

  const logPipelineCR = loadCR('./log-pipeline.yaml');

  it('Operator should be ready', async function() {
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

  describe('Set up mockserver', function() {
    before(async function() {
      await waitForDeployment('mockserver', mockNamespace);
      const {body} = await k8sCoreV1Api.listNamespacedPod(mockNamespace);
      const mockPod = body.items[0].metadata.name;
      await sleep(5000);
      cancelPortForward = kubectlPortForward(
          mockNamespace,
          mockPod,
          mockServerPort,
      );
    });

    it('Should not receive HTTP traffic', async function() {
      await checkMockserverWasCalled(false);
    });

    it('Apply HTTP output plugin to fluent-bit', async function() {
      await k8sApply(logPipelineCR, telemetryNamespace);
      await waitForDaemonSet(fluentBitName, telemetryNamespace, 10000);
    });

    it('Should receive HTTP traffic from fluent-bit', async function() {
      await checkMockserverWasCalled(true);
    });

    after(async function() {
      cancelPortForward();
      await k8sDelete(logPipelineCR, telemetryNamespace);
      await k8sCoreV1Api.deleteNamespace(mockNamespace);
    });
  });
});
