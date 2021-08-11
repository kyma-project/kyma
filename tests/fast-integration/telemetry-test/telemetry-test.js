const uuid = require("uuid");
const k8s = require("@kubernetes/client-node");
const { assert } = require("chai");
const fs = require("fs");
const path = require("path");
const axios = require("axios");
const http = require("http");
const {
  waitForPodWithLabel,
  k8sCoreV1Api,
  k8sCustomApi,
  k8sApply,
  wait,
} = require("../utils");
var ServerMock = require("mock-http-server");
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

  it("should do something", async function () {
    const scope = nock("http://localhost:8080")
      .persist()
      .post("/")
      .reply(200, "Ok");
    await axios.post(
      "http://localhost:8080/",
      { msg: "Hello world!" },
      {
        headers: {
          "Content-type": "application/json; charset=UTF-8",
        },
      }
    );
    assert.equal(scope.isDone(), true);
  });
});

// describe("Telemtry operator", () => {
//   var port = 60000;
//   var server = new ServerMock({ host: "localhost", port: port });

//   // check if operator installed
//   let namespace = "kyma-system";
//   // it("Operator should be ready", async function () {
//   //   let res = await k8sCoreV1Api.listNamespacedPod(
//   //     namespace,
//   //     "true",
//   //     undefined,
//   //     undefined,
//   //     undefined,
//   //     "control-plane=controller-manager,service.istio.io/canonical-name=telemetry-operator-controller-manager"
//   //   );
//   //   let podList = res.body.items;
//   //   assert.equal(podList.length, 1);
//   // });

//   it("Should not receive HTTP traffic", async () => {
//     // server.on({
//     //   method: "POST",
//     //   path: "/",
//     //   reply: {
//     //     status: 200,
//     //     headers: { "content-type": "application/json" },
//     //     body: {
//     //       id: 987654321,
//     //       name: "someName",
//     //       someOtherValue: 1234,
//     //     },
//     //   },
//     // });
//     server.on({
//       method: "GET",
//       path: "/resource",
//       reply: {
//         status: 200,
//         headers: { "content-type": "application/json" },
//         body: JSON.stringify({ hello: "world" }),
//       },
//     });
//     // const options = {
//     //   hostname: 'localhost',
//     //   port: 8000,
//     //   path: '/',
//     //   method: 'POST',
//     //   headers: {
//     //     'Content-Type': 'application/json',
//     //     'Content-Length': data.length
//     //   }
//     // axios
//     //   .post(
//     //     "https://localhost:8080/",
//     //     { msg: "Hello world!" },
//     //     {
//     //       headers: {
//     //         "Content-type": "application/json; charset=UTF-8",
//     //       },
//     //     }
//     //   )
//     //   .then(function (response) {
//     //     console.log(response);
//     //   });
//     axios.get("/resource", { port: port }).then(function (response) {
//       console.log(response);
//     });
//     // var client = https.get(`https://localhost:${port}/resource`, (res) => {
//     //   let data = "";

//     //   res.on("data", (d) => {
//     //     data += d;
//     //   });
//     //   res.on("end", () => {
//     //     console.log(data);
//     //   });
//     // });
//     // https.post("https://localhost:8000");
//     console.log(server.requests());
//   });

//   // it("Create CRD for fluent-bit config", async () => {
//   //   const loggingConfigYaml = fs.readFileSync(
//   //     path.join(__dirname, "./logging-config.yaml"),
//   //     {
//   //       encoding: "utf8",
//   //     }
//   //   );
//   //   const crd = k8s.loadAllYaml(loggingConfigYaml);
//   //   let res = await k8sApply(crd, namespace);

//   //   // console.log(res);

//   //   // let body = {};
//   //   // k8sCustomApi.createNamespacedCustomObject(
//   //   //   "telemetry.kyma-project.io",
//   //   //   "v1",
//   //   //   namespace,
//   //   //   "loggingconfigurations"
//   //   // );
//   // });
// });
