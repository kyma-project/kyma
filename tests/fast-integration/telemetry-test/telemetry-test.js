const k8s = require('@kubernetes/client-node');
const {assert} = require('chai');
const fs = require('fs');
const path = require('path');
const helm = require('./helm');
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

// NOTE: Test needs to be manually triggered at the moment

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

function loadCRD(filepath) {
  const _loggingConfigYaml = fs.readFileSync(path.join(__dirname, filepath), {
    encoding: 'utf8',
  });
  return k8s.loadAllYaml(_loggingConfigYaml);
}

function checkMockserverWasCalled(wasCalled) {
  const args = wasCalled ? [1] : [0, 0];
  const not = wasCalled ? 'not' : '';
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
            assert.fail(`"The HTTP endpoint was ${not} called`);
          },
      );
}

describe('Telemetry operator', function() {
  const telemetryNamespace = 'kyma-system'; // operator flag 'fluent-bit-ns' is set to kyma-system
  const mockNamespace = 'mockserver';
  let cancelPortForward = null;
  const fluentBitName = 'telemetry-fluent-bit';

  const loggingConfigCRD = loadCRD('./logging-config.yaml');

  it('Should install the operator', async () => {
    await helm.installChart(
        'telemetry',
        '../../resources/telemetry',
        telemetryNamespace,
    );
    await waitForDeployment(
        'telemetry-operator-controller-manager',
        telemetryNamespace,
    );
  });

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
      try {
        await k8sCoreV1Api.createNamespace({
          metadata: {name: mockNamespace},
        });
      } catch (error) {
        console.log(`Namespace ${mockNamespace} could not be created`, error);
      }
      await helm.installChart(
          'mockserver',
          './telemetry-test/helm/mockserver',
          mockNamespace,
      );
      await helm.installChart(
          'mockserver-config',
          './telemetry-test/helm/mockserver-config',
          mockNamespace,
      );
      await waitForDeployment('mockserver', mockNamespace);
      const {body} = await k8sCoreV1Api.listNamespacedPod(mockNamespace);
      const mockPod = body.items[0].metadata.name;
      cancelPortForward = kubectlPortForward(
          mockNamespace,
          mockPod,
          mockServerPort,
      );
    });
    after(async function() {
      cancelPortForward();
      await helm.uninstallChart('mockserver', mockNamespace);
      await helm.uninstallChart('mockserver-config', mockNamespace);
      await helm.uninstallChart('telemetry', telemetryNamespace);
      await k8sCoreV1Api.deleteNamespace(mockNamespace);
      k8sDelete(loggingConfigCRD, telemetryNamespace);
    });

    it('Should not receive HTTP traffic', function() {
      return checkMockserverWasCalled(false);
    }).timeout(5000);

    it('Apply HTTP output plugin to fluent-bit', async function() {
      await k8sApply(loggingConfigCRD, telemetryNamespace);
      await sleep(10000); // wait for controller to reconcile
      await waitForDaemonSet(fluentBitName, telemetryNamespace);
    });

    it('Should receive HTTP traffic from fluent-bit', function() {
      return checkMockserverWasCalled(true);
    }).timeout(5000);
  });
});
