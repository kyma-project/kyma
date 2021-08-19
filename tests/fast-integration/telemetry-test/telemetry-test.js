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
} = require("../utils");
const mockServerClient = require("mockserver-client").mockServerClient;

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

describe("Telemetry operator", function () {
  let namespace = "kyma-system";
  let mockNamespace = "mockserver";
  let mockServerPort = 1080;
  let cancelPortForward = null;

  const _loggingConfigYaml = fs.readFileSync(
    path.join(__dirname, "./logging-config.yaml"),
    {
      encoding: "utf8",
    }
  );
  const loggingConfigCRD = k8s.loadAllYaml(_loggingConfigYaml);

  it("Operator should be ready", async function () {
    let res = await k8sCoreV1Api.listNamespacedPod(
      namespace,
      "true",
      undefined,
      undefined,
      undefined,
      "control-plane=controller-manager,service.istio.io/canonical-name=telemetry-operator-controller-manager"
    );
    let podList = res.body.items;
    assert.equal(podList.length, 1);
  });
  describe("Set up mockserver", function () {
    before(async function () {
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
      await waitForDeployment("mockserver", "mockserver");
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
      k8sDelete(loggingConfigCRD, namespace);
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
      await k8sApply(loggingConfigCRD, namespace);
      await sleep(10000); // wait for controller to reconcile
      await waitForDaemonSet("logging-fluent-bit", namespace);
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
  }).timeout(80000);
});
