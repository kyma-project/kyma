const uuid = require("uuid");
const k8s = require("@kubernetes/client-node");
const { assert } = require("chai");
const fs = require("fs");
const path = require("path");
const axios = require("axios");
const {
  waitForPodWithLabel,
  k8sCoreV1Api,
  k8sDynamicApi,
  k8sApply,
  k8sDelete,
  kubectlPortForward,
} = require("../utils");
const nock = require("nock");

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}
const mockServerClient = require("mockserver-client").mockServerClient;

describe("Telemtry operator", () => {
  // var port = 8080;
  var server; // TODO

  let namespace = "kyma-system";
  let mockNamespace = "mockserver";
  let mockServerPort = 9999;

  const loggingConfigYaml = fs.readFileSync(
    path.join(__dirname, "./logging-config.yaml"),
    {
      encoding: "utf8",
    }
  );
  const loggingConfigCRD = k8s.loadAllYaml(loggingConfigYaml);

  after(() => {
    // delete custom config
    k8sDelete(loggingConfigCRD, namespace);
  });

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

  // it("should not receive HTTP traffic", async function () {
  //   assert.equal(server.isDone(), false);
  // });

  // it("should receive HTTP traffic", async function () {
  //   await axios.post(
  //     `http://localhost:${port}`,
  //     { msg: "Hello world!" },
  //     {
  //       headers: {
  //         "Content-type": "application/json; charset=UTF-8",
  //       },
  //     }
  //   );
  //   assert.equal(server.isDone(), true);
  // });

  describe("Prepare mockserver", function () {
    before(() => {
      // install helm chart
      // configure port forward..
    });
    after(() => {
      // uninstall helm chart
    });

    it("Should not receive HTTP traffic", async function () {
      //
    });

    it("Apply HTTP output plugin to fluent-bit", async function () {
      // wait for pod restart
    });

    it("Should receive HTTP traffic from fluent-bit", async function () {
      // verify server
    });
  });

  it("Should not receive HTTP traffic", async function () {
    let { body } = await k8sCoreV1Api.listNamespacedPod(mockNamespace);
    let mockPod = body.items[0].name;
    kubectlPortForward(mockNamespace, mockPod, mockServerPort);
    mockServerClient("localhost", mockServerPort)
      .verify(
        {
          path: "/",
        },
        0,
        true
      )
      .then(
        function () {
          console.log("request found exactly 0 times");
        },
        function (error) {
          // throw error;
          console.log("Not exactly 0 times error");
        }
      );
  });

  it("Create CRD for fluent-bit config", async () => {
    let res = await k8sApply(loggingConfigCRD, namespace);
    // kubectlPortForward(namespace, "logging-fluent-bit-5qvxz", port);

    // nock.recorder.rec();
    // pname = "logging-fluent-bit-5qvxz";
    let { body } = await k8sCoreV1Api.listNamespacedPod(mockNamespace);
    let mockPod = body.items[0].name;
    // kubectlPortForward(mockNamespace, mockPod, mockServerPort);
    mockServerClient("localhost", mockServerPort)
      .verify(
        {
          path: "/",
        },
        1
      )
      .then(
        function () {
          console.log("request found 2 times");
        },
        function (error) {
          // throw error;
          // console.log(error);
        }
      );
    // console.log(reso);
    // let pod = await k8sCoreV1Api.createNamespacedPod(namespace, {
    //   metadata: { name: "test-server" },
    //   spec: {
    //     containers: [
    //       {
    //         name: "server",
    //         image: "mockserver/mockserver",
    //         ports: [
    //           {
    //             name: "server-port",
    //             containerPort: 8081,
    //           },
    //         ],
    //       },
    //     ],
    //   },
    // });
    // console.log("path", k8sDynamicApi.basePath);
    // let { body } = await k8sDynamicApi.requestPromise({
    //   url:
    //     k8sDynamicApi.basePath +
    //     `/api/v1/namespaces/${namespace}/pods/${pname}/log`,
    //   method: "GET",
    // });
    // let { body } = await k8sCoreV1Api.readNamespacedPodLog(
    //   pname,
    //   namespace,
    //   undefined,
    //   undefined,
    //   true,
    //   1000,
    //   "false",
    //   false,
    //   undefined,
    //   undefined,
    //   undefined,
    //   undefined
    // );
    // console.log(body);
    // await axios.post(
    //   `http://localhost:${port}`,
    //   { msg: "Hello world!" },
    //   {
    //     headers: {
    //       "Content-type": "application/json; charset=UTF-8",
    //     },
    //   }
    // );
    // await sleep(10000);

    // assert.equal(server.isDone(), true);
  }).timeout(20000);
});
