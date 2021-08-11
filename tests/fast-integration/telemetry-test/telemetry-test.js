const uuid = require("uuid");
const k8s = require("@kubernetes/client-node");
const { assert } = require("chai");
const fs = require("fs");
const path = require("path");
const axios = require("axios");
const { waitForPodWithLabel, k8sCoreV1Api, k8sApply } = require("../utils");
const nock = require("nock");

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

describe("Telemtry operator", () => {
  var port = 8080;
  var server;

  // check if operator installed
  let namespace = "kyma-system";

  beforeEach(function () {
    server = nock(`http://localhost:${port}`)
      .persist()
      .post("/")
      .reply(200, "Ok");
  });

  afterEach(function () {
    nock.cleanAll();
  });

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

  it("should not receive HTTP traffic", async function () {
    assert.equal(server.isDone(), false);
  });

  it("should receive HTTP traffic", async function () {
    await axios.post(
      `http://localhost:${port}`,
      { msg: "Hello world!" },
      {
        headers: {
          "Content-type": "application/json; charset=UTF-8",
        },
      }
    );
    assert.equal(server.isDone(), true);
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

    nock.recorder.rec();
    // await axios.post(
    //   `http://localhost:${port}`,
    //   { msg: "Hello world!" },
    //   {
    //     headers: {
    //       "Content-type": "application/json; charset=UTF-8",
    //     },
    //   }
    // );
    await sleep(10000);

    assert.equal(server.isDone(), true);
  }).timeout(20000);
});
