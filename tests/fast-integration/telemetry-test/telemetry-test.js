const uuid = require("uuid");

const { assert } = require("chai");

const { waitForPodWithLabel, k8sCoreV1Api } = require("../utils");

describe("Telemtry operator", function () {
  // check if operator installed
  it("Operator should be ready", async function () {
    let namespace = "kyma-system";
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

  it;
});
