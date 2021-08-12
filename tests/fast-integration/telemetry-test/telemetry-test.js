const uuid = require("uuid");
const k8s = require("@kubernetes/client-node");
const { assert } = require("chai");
const fs = require("fs");
const path = require("path");
const axios = require("axios");
const {
  waitForPodWithLabel,
  waitForDaemonSet,
  k8sCoreV1Api,
  k8sDynamicApi,
  k8sApply,
  k8sDelete,
  kubectlPortForward,
} = require("../utils");

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}
const mockServerClient = require("mockserver-client").mockServerClient;

describe("Telemtry operator", () => {
  let namespace = "kyma-system";
  let mockNamespace = "mockserver";
  let mockServerPort = 1080;

  const loggingConfigYaml = fs.readFileSync(
    path.join(__dirname, "./logging-config.yaml"),
    {
      encoding: "utf8",
    }
  );
  const loggingConfigCRD = k8s.loadAllYaml(loggingConfigYaml);

  after(() => {
    // delete custom config TODO
    k8sDelete(loggingConfigCRD, namespace);
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
  describe("Prepare mockserver", function () {
    before(async () => {
      // install helm chart
      let { body } = await k8sCoreV1Api.listNamespacedPod(mockNamespace);
      let mockPod = body.items[0].metadata.name;
      kubectlPortForward(mockNamespace, mockPod, mockServerPort); //forward service?

      // wait for pod to be ready
    });
    after(() => {
      // uninstall helm chart
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
    });

    it("Apply HTTP output plugin to fluent-bit", async function () {
      await k8sApply(loggingConfigCRD, namespace);
      // await waitForDaemonSet("logging-fluent-bit", namespace);// TODO
      await sleep(70000);
    }).timeout(90000);

    it("Should receive HTTP traffic from fluent-bit", function () {
      // verify server
      return mockServerClient("localhost", mockServerPort)
        .verify(
          {
            path: "/",
          },
          1
        )
        .then(
          function () {
            console.log("request found 1 times");
          },
          function (error) {
            assert.fail("The HTTP endpoint was not called");
          }
        );
    });
  });

  // it("Create CRD for fluent-bit config", async () => {
  //   let res = await k8sApply(loggingConfigCRD, namespace);
  //   let { body } = await k8sCoreV1Api.listNamespacedPod(mockNamespace);
  //   let mockPod = body.items[0].name;
  //   // kubectlPortForward(mockNamespace, mockPod, mockServerPort);
  //   mockServerClient("localhost", mockServerPort)
  //     .verify(
  //       {
  //         path: "/",
  //       },
  //       1
  //     )
  //     .then(
  //       function () {
  //         console.log("request found 2 times");
  //       },
  //       function (error) {
  //         // throw error;
  //         // console.log(error);
  //       }
  //     );

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
  // }).timeout(20000);
});
