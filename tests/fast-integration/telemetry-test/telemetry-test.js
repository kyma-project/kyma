const k8s = require("@kubernetes/client-node");
const { assert } = require("chai");
const fs = require("fs");
const path = require("path");
const helm = require("./helm");
const {
  waitForDaemonSet,
  waitForPodWithLabel,
  waitForDeployment,
  k8sCoreV1Api,
  k8sApply,
  k8sDelete,
  kubectlPortForward,
  debug,
} = require("../utils");
const mockServerClient = require("mockserver-client").mockServerClient;

let mockServerPort = 1080;

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

function loadCRD(filepath) {
  const _loggingConfigYaml = fs.readFileSync(path.join(__dirname, filepath), {
    encoding: "utf8",
  });
  return k8s.loadAllYaml(_loggingConfigYaml);
}

function checkMockserverWasCalled(wasCalled) {
  const args = wasCalled ? [1] : [0, 0];
  const not = wasCalled ? "not" : "";
  return mockServerClient("localhost", mockServerPort)
    .verify(
      {
        path: "/",
      },
      ...args
    )
    .then(
      function () {},
      function (error) {
        assert.fail(`"The HTTP endpoint was ${not} called`);
      }
    );
}

describe("Telemetry operator", function () {
  let telemetryNamespace = "kubesphere-logging-system"; // operator flag 'fluent-bit-ns' is set to kyma-system
  let mockNamespace = "mockserver";
  let cancelPortForward = null;
  let fluentBitName = "fluent-bit";

  const loggingConfigCRD = loadCRD("./logging-config.yaml");

  it("Should install the operator", async () => {
    try {
      await k8sCoreV1Api.createNamespace({
        metadata: { name: mockNamespace },
      });
    } catch (error) {
      console.log(`Namespace ${telemetryNamespace} could not be created`);
    }
    await k8sApply(loadCRD("./setup-operator.yaml"));
  });

  it("Operator should be ready", async function () {
    await waitForDeployment("fluentbit-operator", telemetryNamespace);
    await k8sApply(loadCRD("./install-fb.yaml"));
    // await waitForDaemonSet(fluentBitName, telemetryNamespace);
    await waitForPodWithLabel(
      "app.kubernetes.io/name",
      "fluent-bit",
      telemetryNamespace
    );
  });
  describe("Set up mockserver", function () {
    before(async function () {
      try {
        await k8sCoreV1Api.createNamespace({
          metadata: { name: mockNamespace },
        });
      } catch (error) {
        console.log(`Namespace ${mockNamespace} could not be created`);
      }
      await helm.installChart(
        "mockserver",
        "./telemetry-test/helm/mockserver",
        mockNamespace
      );
      await helm.installChart(
        "mockserver-config",
        "./telemetry-test/helm/mockserver-config",
        mockNamespace
      );
      await waitForDeployment("mockserver", mockNamespace);
      let { body } = await k8sCoreV1Api.listNamespacedPod(mockNamespace);
      let mockPod = body.items[0].metadata.name;
      cancelPortForward = kubectlPortForward(
        mockNamespace,
        mockPod,
        mockServerPort
      );
    });
    after(async function () {
      cancelPortForward();
      await helm.uninstallChart("mockserver", mockNamespace);
      await helm.uninstallChart("mockserver-config", mockNamespace);
      await helm.uninstallChart("telemetry", telemetryNamespace);
      await k8sCoreV1Api.deleteNamespace(mockNamespace);
      await k8sDelete(loggingConfigCRD, telemetryNamespace);
      await k8sDelete(loadCRD("./install-fb.yaml"), telemetryNamespace);
      await k8sDelete(loadCRD("./setup-operator.yaml"), telemetryNamespace);
    });

    it("Should not receive HTTP traffic", function () {
      return checkMockserverWasCalled(false);
    }).timeout(5000);

    it("Apply HTTP output plugin to fluent-bit", async function () {
      await k8sApply(loggingConfigCRD, telemetryNamespace);
      await sleep(600000); // wait for controller to reconcile
    });

    it("Should receive HTTP traffic from fluent-bit", function () {
      return checkMockserverWasCalled(true);
    }).timeout(5000);
  });
});
