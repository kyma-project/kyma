import { defaults } from "axios";
import { Agent } from "https";
import { ensureCommerceMockLocalTestFixture, checkFunctionResponse, sendEventAndCheckResponse, cleanMockTestFixture, checkInClusterEventDelivery } from "./fixtures/commerce-mock";
import { printRestartReport, getContainerRestartsForAllNamespaces } from "../utils";
import { checkLokiLogs, lokiPortForward } from "../logging";

const httpsAgent = new Agent({
  rejectUnauthorized: false, // curl -k
});
defaults.httpsAgent = httpsAgent;

describe("CommerceMock tests", function () {
  this.timeout(10 * 60 * 1000);
  this.slow(5000);
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

  it("CommerceMock test fixture should be ready", async function () {
    await ensureCommerceMockLocalTestFixture("mocks", testNamespace).catch((err) => {
      console.dir(err); // first error is logged
      return ensureCommerceMockLocalTestFixture("mocks", testNamespace);
    });
  });

  it("in-cluster event should be delivered", async function () {
    await checkInClusterEventDelivery(testNamespace);
  });

  it("function should be reachable through secured API Rule", async function () {
    await checkFunctionResponse(testNamespace);
  });

  it("order.created.v1 event should trigger the lastorder function", async function () {
    await sendEventAndCheckResponse();
  });

  it("Should print report of restarted containers, skipped if no crashes happened", async function () {
    const afterTestRestarts = await getContainerRestartsForAllNamespaces();
    printRestartReport(initialRestarts, afterTestRestarts);
  });

  it("Logs from commerce mock pod should be retrieved through Loki", async function() {
    await checkLokiLogs(testStartTimestamp);
  });

  it("Test namespaces should be deleted", async function () {
    await cleanMockTestFixture("mocks", testNamespace, true);
  });
});
