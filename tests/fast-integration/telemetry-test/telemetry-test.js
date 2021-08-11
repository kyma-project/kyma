const uuid = require("uuid");
const k8s = require("@kubernetes/client-node");
const { assert } = require("chai");
const fs = require("fs");
const path = require("path");

const {
  waitForPodWithLabel,
  k8sCoreV1Api,
  k8sCustomApi,
  k8sApply,
} = require("../utils");

describe("Telemtry operator", () => {
  // check if operator installed
  let namespace = "kyma-system";
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

  it("Create CRD for fluent-bit config", async () => {
    const loggingConfigYaml = fs.readFileSync(
      path.join(__dirname, "./logging-config.yaml"),
      {
        encoding: "utf8",
      }
    );
    const crd = k8s.loadAllYaml(loggingConfigYaml);
    let res = await k8sApply(crd, namespace);

    console.log(res);

    // let body = {};
    // k8sCustomApi.createNamespacedCustomObject(
    //   "telemetry.kyma-project.io",
    //   "v1",
    //   namespace,
    //   "loggingconfigurations"
    // );
  });
});
