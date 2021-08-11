const uuid = require("uuid");
const k8s = require("@kubernetes/client-node");
const { assert } = require("chai");
const fs = require("fs");
const path = require("path");
const axios = require("axios");
const {
  waitForPodWithLabel,
  k8sCoreV1Api,
  k8sCustomApi,
  k8sApply,
  wait,
} = require("../utils");
const nock = require("nock");
// beforeEach(function (done) {
//   server.start(done);
//   console.log(server.getHttpPort());
// });

// afterEach(function (done) {
//   server.stop(done);
// });

describe("Test", async function () {
  // http.get("http://localhost/"); // respond body "Ok"
  // http.get("http://localhost/"); // respond body "Ok"
  // http.get("http://localhost/"); // respond body "Ok"
  // Run an HTTP server on localhost:8080
  // var server = new ServerMock({ host: "localhost", port: 8080 });
  // beforeEach(function (done) {
  //   server.start(done);
  // });
  // afterEach(function (done) {
  //   server.stop(done);
  // });
});

describe("Telemtry operator", () => {
  var port = 8080;
  var server;

  // check if operator installed
  let namespace = "kyma-system";
  // it("Operator should be ready", async function () {
  //   let res = await k8sCoreV1Api.listNamespacedPod(
  //     namespace,
  //     "true",
  //     undefined,
  //     undefined,
  //     undefined,
  //     "control-plane=controller-manager,service.istio.io/canonical-name=telemetry-operator-controller-manager"
  //   );
  //   let podList = res.body.items;
  //   assert.equal(podList.length, 1);
  // });

  beforeEach(function () {
    server = nock(`http://localhost:${port}`)
      .persist()
      .post("/")
      .reply(200, "Ok");
  });

  afterEach(function () {
    nock.cleanAll();
  });

  it("should not receive HTTP traffic", async function () {
    assert.equal(server.isDone(), false);
  });

  it("should receive HTTP traffic", async function () {
    let res = await axios.post(
      `http://localhost:${port}`,
      { msg: "Hello world!" },
      {
        headers: {
          "Content-type": "application/json; charset=UTF-8",
        },
      }
    );
    // console.log(res);
    assert.equal(server.isDone(), true);
  });

  // it("Create CRD for fluent-bit config", async () => {
  //   const loggingConfigYaml = fs.readFileSync(
  //     path.join(__dirname, "./logging-config.yaml"),
  //     {
  //       encoding: "utf8",
  //     }
  //   );
  //   const crd = k8s.loadAllYaml(loggingConfigYaml);
  //   let res = await k8sApply(crd, namespace);

  //   // console.log(res);

  //   // let body = {};
  //   // k8sCustomApi.createNamespacedCustomObject(
  //   //   "telemetry.kyma-project.io",
  //   //   "v1",
  //   //   namespace,
  //   //   "loggingconfigurations"
  //   // );
  // });
});
