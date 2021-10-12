const axios = require("axios");
const https = require("https");
const httpsAgent = new https.Agent({
  rejectUnauthorized: false, // curl -k
});
axios.defaults.httpsAgent = httpsAgent;
const {
  ensureCommerceMockLocalTestFixture,
  checkFunctionResponse,
  sendEventAndCheckResponse,
  cleanMockTestFixture,
  checkInClusterEventDelivery,
  waitForSubscriptionsTillReady,
  setEventMeshSourceNamespace,
} = require("../test/fixtures/commerce-mock");
const {
  switchEventingBackend,
  createEventingBackendK8sSecret,
  deleteEventingBackendK8sSecret,
  printAllSubscriptions,
  printEventingControllerLogs,
  printEventingPublisherProxyLogs,
} = require("../utils");

describe("Eventing tests", function () {
  this.timeout(10 * 60 * 1000);
  this.slow(5000);
  const testNamespace = "test";
  const backendK8sSecretName = process.env.BACKEND_SECRET_NAME || "eventing-backend";
  const backendK8sSecretNamespace = process.env.BACKEND_SECRET_NAMESPACE || "default";
  const eventMeshSecretFilePath = process.env.EVENTMESH_SECRET_FILE || "";
  const DEBUG = process.env.DEBUG;

  // eventingE2ETestSuite - Runs Eventing end-to-end tests
  function eventingE2ETestSuite () {
    it("lastorder function should be reachable through secured API Rule", async function () {
      await checkFunctionResponse(testNamespace);
    });

    it("In-cluster event should be delivered (structured and binary mode)", async function () {
      await checkInClusterEventDelivery(testNamespace);
    });

    it("order.created.v1 event from CommerceMock should trigger the lastorder function", async function () {
      await sendEventAndCheckResponse();
    });
  }

  before(async function() {
    // runs once before the first test in this block

    // If eventMeshSecretFilePath is specified then create a k8s secret for eventing-backend
    // else use existing k8s secret as specified in backendK8sSecretName & backendK8sSecretNamespace
    if (eventMeshSecretFilePath !== "") {
      console.log("Creating Event Mesh secret")
      const eventMeshInfo = await createEventingBackendK8sSecret(eventMeshSecretFilePath, backendK8sSecretName, backendK8sSecretNamespace);
      setEventMeshSourceNamespace(eventMeshInfo["namespace"]);
    }

    // Deploy Commerce mock application, function and subscriptions for tests
    console.log("Preparing CommerceMock test fixture")
    await ensureCommerceMockLocalTestFixture("mocks", testNamespace).catch((err) => {
      console.dir(err); // first error is logged
      return ensureCommerceMockLocalTestFixture("mocks", testNamespace);
    });
  });

  after(async function() {
    // runs once after the last test in this block
    console.log("Cleaning: Test namespaces should be deleted")
    await cleanMockTestFixture("mocks", testNamespace, true);

    // Delete eventing backend secret if it was created by test
    if (eventMeshSecretFilePath !== "") {
      await deleteEventingBackendK8sSecret(backendK8sSecretName, backendK8sSecretNamespace);
    }
  });

  afterEach(async function() {
    // runs after each test in every block

    // if the test is failed, then printing some debug logs
    if (this.currentTest.state === 'failed' && DEBUG) {
      await printAllSubscriptions(testNamespace)
      await printEventingControllerLogs()
      await printEventingPublisherProxyLogs()
    }
  });

  // Tests
  context('with Nats backend', function() {
    // Running Eventing end-to-end tests
    eventingE2ETestSuite();
  });

  context('with BEB backend', function() {
    it("Switch Eventing Backend to BEB", async function () {
      await switchEventingBackend(backendK8sSecretName, backendK8sSecretNamespace, "beb");
      await waitForSubscriptionsTillReady(testNamespace)

      // print subscriptions status when debugLogs is enabled
      if (DEBUG) {
        await printAllSubscriptions(testNamespace)
      }
    });

    // Running Eventing end-to-end tests
    eventingE2ETestSuite();
  });

  context('with Nats backend switched back from BEB', function() {
    it("Switch Eventing Backend to Nats", async function () {
      await switchEventingBackend(backendK8sSecretName, backendK8sSecretNamespace, "nats");
      await waitForSubscriptionsTillReady(testNamespace)
    });

    // Running Eventing end-to-end tests
    eventingE2ETestSuite();
  });
});
