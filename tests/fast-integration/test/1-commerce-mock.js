const axios = require("axios");
const https = require("https");
const httpsAgent = new https.Agent({
  rejectUnauthorized: false, // curl -k
});
axios.defaults.httpsAgent = httpsAgent;
const {
  ensureCommerceMockLocalTestFixture,
  checkFunctionResponse,
  addService,
  updateService,
  deleteService,
  checkRevocation,
  sendEventAndCheckResponse,
  renewCommerceMockCertificate,
  revokeCommerceMockCertificate,
  cleanMockTestFixture,
  checkInClusterEventDelivery,
} = require("./fixtures/commerce-mock");
const {
  printRestartReport,
  getContainerRestartsForAllNamespaces,
} = require("../utils");
const { checkLokiLogs, lokiPortForward } = require("../logging");

function commerceMockTests() {
  describe("CommerceMock tests", function () {
    this.timeout(10 * 60 * 1000);
    this.slow(5000);
    const withCentralAppConnectivity =
      process.env.WITH_CENTRAL_APP_CONNECTIVITY === "true";
    const testNamespace = "test";
    const testStartTimestamp = new Date().toISOString();
    let initialRestarts = null;
    let cancelPortForward = null;

    before(() => {
      cancelPortForward = lokiPortForward();
    });

    after(() => {
      cancelPortForward();
    });

    it("Listing all pods in cluster", async function () {
      initialRestarts = await getContainerRestartsForAllNamespaces();
    });

    /*it("Wait 60 min to check the logs", async function () {
      await new Promise(resolve => setTimeout(resolve, 600000));
    });*/

    it("CommerceMock test fixture should be ready", async function () {
      await ensureCommerceMockLocalTestFixture(
        "mocks",
        testNamespace,
        withCentralAppConnectivity
      ).catch((err) => {
        console.dir(err); // first error is logged
        return ensureCommerceMockLocalTestFixture(
          "mocks",
          testNamespace,
          withCentralAppConnectivity
        );
      });
    });

    it("in-cluster event should be delivered (structured and binary mode)", async function () {
      await checkInClusterEventDelivery(testNamespace);
    });

    it("function should be reachable through secured API Rule", async function () {
      await checkFunctionResponse(testNamespace);
    });

    it("order.created.v1 event should trigger the lastorder function", async function () {
      await sendEventAndCheckResponse();
    });

    it("should add, update and delete a service", async function () {
      let serviceId = await addService();
      await updateService(serviceId);
      await deleteService(serviceId);
    });

    it("CommerceMock should renew it's certificate", async function () {
      await renewCommerceMockCertificate();
    });

    it("the event should triger the fucntion after the certificate renewal", async function () {
      await sendEventAndCheckResponse();
    });

    it("should revoke Commerce Mock certificate", async function () {
      await revokeCommerceMockCertificate();
    });

    it("the event should triger the fucntion after revoke of the certificate", async function () {
      await sendEventAndCheckResponse();
    });

    it("should pass if the certificated is revoked, endpoint returned 403 code", async function () {
      await checkRevocation();
    });

    it("Should print report of restarted containers, skipped if no crashes happened", async function () {
      const afterTestRestarts = await getContainerRestartsForAllNamespaces();
      printRestartReport(initialRestarts, afterTestRestarts);
    });

    it("Logs from commerce mock pod should be retrieved through Loki", async function () {
      await checkLokiLogs(testStartTimestamp);
    });

    it("Test namespaces should be deleted", async function () {
      await cleanMockTestFixture("mocks", testNamespace, true);
    });
  });
}

module.exports = {
  commerceMockTests,
};
