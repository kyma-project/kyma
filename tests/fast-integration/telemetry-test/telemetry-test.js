const k8s = require("@kubernetes/client-node");
const { assert } = require("chai");
const fs = require("fs");
const path = require("path");
const helm = require("./helm");
const {
  waitForDaemonSet,
  waitForDeployment,
  k8sCoreV1Api,
  k8sApply,
  k8sDelete,
  kubectlPortForward,
  debug,
} = require("../utils");
const mockServerClient = require("mockserver-client").mockServerClient;

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

function loadCRD(filepath) {
  const _loggingConfigYaml = fs.readFileSync(path.join(__dirname, filepath), {
    encoding: "utf8",
  });
  return k8s.loadAllYaml(_loggingConfigYaml);
}

describe("Telemetry operator", function () {
  let telemetryNamespace = "kyma-system"; // operator flag 'fluent-bit-ns' is set to kyma-system
  let mockNamespace = "mockserver";
  let mockServerPort = 1080;
  let cancelPortForward = null;
  let fluentBitName = "telemetry-fluent-bit";

  const loggingConfigCRD = loadCRD("./logging-config.yaml");

  it("Should install the operator", async () => {
    // await k8sCoreV1Api.createNamespace({
    //   metadata: { name: telemetryNamespace },
    // });
    await helm.installChart(
      "telemetry",
      "../../resources/telemetry",
      telemetryNamespace
    );
    await waitForDeployment(
      "telemetry-operator-controller-manager",
      telemetryNamespace
    );
  });

  it("Operator should be ready", async function () {
    let res = await k8sCoreV1Api.listNamespacedPod(
      telemetryNamespace,
      "true",
      undefined,
      undefined,
      undefined,
      "control-plane=telemetry-operator-controller-manager"
    );
    let podList = res.body.items;
    assert.equal(podList.length, 1);
  });
  describe("Set up mockserver", function () {
    before(async function () {
      try {
        await k8sCoreV1Api.createNamespace({
          metadata: { name: mockNamespace },
        });
      } catch (error) {
        console.log(
          `Namespace ${telemetryNamespace} could not be created`,
          error
        );
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
      await helm.uninstallChart("mockserver", "mockserver");
      await helm.uninstallChart("mockserver-config", "mockserver");
      await k8sCoreV1Api.deleteNamespace(mockNamespace);
      // await k8sCoreV1Api.deleteNamespace(telemetryNamespace);
      k8sDelete(loggingConfigCRD, telemetryNamespace);
    });

    it("Should not receive HTTP traffic", function () {
      return mockServerClient("localhost", mockServerPort)
        .verify(
          {
            path: "/",
          },
          0,
          0
        )
        .then(
          function () {},
          function (error) {
            assert.fail("HTTP endpoint was called");
          }
        );
    }).timeout(5000);

    it("Apply HTTP output plugin to fluent-bit", async function () {
      await k8sApply(loggingConfigCRD, telemetryNamespace);
      await sleep(10000); // wait for controller to reconcile
      await waitForDaemonSet(fluentBitName, telemetryNamespace);
    });

    it("Should receive HTTP traffic from fluent-bit", function () {
      return mockServerClient("localhost", mockServerPort)
        .verify(
          {
            path: "/",
          },
          1
        )
        .then(
          function () {},
          function (error) {
            assert.fail("The HTTP endpoint was not called");
          }
        );
    }).timeout(5000);
  });
});
