const k8s = require("@kubernetes/client-node");
const {
  ensureCommerceMockTestFixture,
  checkAppGatewayResponse,
  sendEventAndCheckResponse,
  cleanMockTestFixture,
} = require("./fixtures/commerce-mock");
const {
  getRestartRaport,
  getContainerRestartsForAllNamespaces,
} = require("../utils");

describe("CommerceMock tests", function () {
  this.timeout(10 * 60 * 1000);
  this.slow(5000);
  const testNamespace = "test";
  let initialRestarts = null;

  it("Listing all pods in cluster", async function () {
    initialRestarts = await getContainerRestartsForAllNamespaces();
  });

  it("CommerceMock test fixture should be ready", async function () {
    await ensureCommerceMockTestFixture("mocks", testNamespace).catch((err) => {
      console.dir(err); // first error is logged
      return ensureCommerceMockTestFixture("mocks", testNamespace);
    });
  });

  it("function should reach Commerce mock API through app gateway", async function () {
    await checkAppGatewayResponse();
  });

  it("order.created.v1 event should trigger the lastorder function", async function () {
    await sendEventAndCheckResponse();
  });

  it("Spit out raport", async function () {
    const afterTestRestarts = await getContainerRestartsForAllNamespaces();
    // console.log(JSON.stringify(asd(initialRestarts, afterTestRestarts)));
    console.log(
      k8s.dumpYaml(getRestartRaport(initialRestarts, afterTestRestarts))
    );
  });

  it("Test namespaces should be deleted", async function () {
    await cleanMockTestFixture("mocks", testNamespace, false);
  });
});
