const axios = require("axios");
const https = require("https");
const { expect, assert } = require("chai");
const httpsAgent = new https.Agent({
  rejectUnauthorized: false, // curl -k
});
axios.defaults.httpsAgent = httpsAgent;
const {
  ensureCommerceMockLocalTestFixture,
  checkFunctionResponse,
  sendEventAndCheckResponse,
  sendLegacyEventAndCheckTracing,
  cleanMockTestFixture,
  checkInClusterEventDelivery,
  checkInClusterEventTracing,
  waitForSubscriptionsTillReady,
  setEventMeshSourceNamespace,
  ensureCommerceMockWithCompassTestFixture,
  cleanCompassResourcesSKR,
} = require("../test/fixtures/commerce-mock");
const {
  debug,
  getShootNameFromK8sServerUrl,
  switchEventingBackend,
  createEventingBackendK8sSecret,
  deleteEventingBackendK8sSecret,
  printAllSubscriptions,
  printEventingControllerLogs,
  printEventingPublisherProxyLogs,
  genRandom,
} = require("../utils");
const {
  DirectorClient,
  DirectorConfig,
  addScenarioInCompass,
  assignRuntimeToScenario,
} = require("../compass");
const {prometheusPortForward} = require("../monitoring/client")
const {eventingMonitoringTest} = require("./metric-test")
const {GardenerClient, GardenerConfig} = require("../gardener");


describe("Eventing tests", function () {
  this.timeout(10 * 60 * 1000);
  this.slow(5000);
  let suffix = genRandom(4);
  const appName = `app-${suffix}`;
  const scenarioName = `test-${suffix}`;
  const testNamespace = `test-${suffix}`;
  const mockNamespace = process.env.MOCK_NAMESPACE || 'mocks'
  const isSKR = process.env.KYMA_TYPE === "SKR";
  const backendK8sSecretName = process.env.BACKEND_SECRET_NAME || "eventing-backend";
  const backendK8sSecretNamespace = process.env.BACKEND_SECRET_NAMESPACE || "default";
  const eventMeshSecretFilePath = process.env.EVENTMESH_SECRET_FILE || "";
  const DEBUG = process.env.DEBUG;
  let cancelPrometheusPortForward = null;
  let gardener = null;
  let director = null;
  let skrInfo = null;

  // eventingE2ETestSuite - Runs Eventing end-to-end tests
  function eventingE2ETestSuite () {
    it("lastorder function should be reachable through secured API Rule", async function () {
      await checkFunctionResponse(testNamespace, mockNamespace);
    });

    it("In-cluster event should be delivered (structured and binary mode)", async function () {
      await checkInClusterEventDelivery(testNamespace);
    });

    it("order.created.v1 event from CommerceMock should trigger the lastorder function", async function () {
      await sendEventAndCheckResponse(mockNamespace);
    });
  }

  // eventingTracingTestSuite - Runs Eventing tracing tests
  function eventingTracingTestSuite () {
    // Only run tracing tests on OSS
    if (isSKR) {
      console.log("Skipping eventing tracing tests on SKR...")
      return
    }

    it("order.created.v1 event from CommerceMock should have correct tracing spans", async function () {
      await sendLegacyEventAndCheckTracing(testNamespace, mockNamespace);
    });
    it("In-cluster event should have correct tracing spans", async function () {
      await checkInClusterEventTracing(testNamespace);
    });
  }

  // prepareAssetsForOSSTests - Sets up CommerceMost for the OSS
  async function prepareAssetsForOSSTests() {
    console.log("Preparing CommerceMock test fixture on Kyma OSS")
    await ensureCommerceMockLocalTestFixture(mockNamespace, testNamespace).catch((err) => {
      console.dir(err); // first error is logged
      return ensureCommerceMockLocalTestFixture(mockNamespace, testNamespace);
    });
  }

  // prepareAssetsForSKRTests - Sets up CommerceMost for the SKR
  async function prepareAssetsForSKRTests() {
    console.log("Preparing for tests on SKR")
    // create gardener & director clients
    gardener = new GardenerClient(GardenerConfig.fromEnv());
    // director client for Compass
    director = new DirectorClient(DirectorConfig.fromEnv());

    // Get shoot info from gardener to get compassID for this shoot
    const shootName = getShootNameFromK8sServerUrl();
    console.log(`Fetching SKR info for shoot: ${shootName}`)
    skrInfo = await gardener.getShoot(shootName)
    debug(`appName: ${appName}, scenarioName: ${scenarioName}, testNamespace: ${testNamespace}, compassID: ${skrInfo.compassID}`);

    console.log("Assigning SKR to scenario in Compass")
    // Create a new scenario (systems/formations) in compass for this test
    await addScenarioInCompass(director, scenarioName);
    // map scenario to target SKR
    await assignRuntimeToScenario(director, skrInfo.compassID, scenarioName);

    console.log("Preparing CommerceMock test fixture on Kyma SKR")
    await ensureCommerceMockWithCompassTestFixture(
        director,
        appName,
        scenarioName,
        mockNamespace,
        testNamespace
    );
  }

  before(async function() {
    // runs once before the first test in this block
    console.log('Running with mockNamspace =', mockNamespace)

    // If eventMeshSecretFilePath is specified then create a k8s secret for eventing-backend
    // else use existing k8s secret as specified in backendK8sSecretName & backendK8sSecretNamespace
    if (eventMeshSecretFilePath !== "") {
      console.log("Creating Event Mesh secret")
      const eventMeshInfo = await createEventingBackendK8sSecret(eventMeshSecretFilePath, backendK8sSecretName, backendK8sSecretNamespace);
      setEventMeshSourceNamespace(eventMeshInfo["namespace"]);
    }

    // Deploy Commerce mock application, function and subscriptions for tests
    if (isSKR) {
      await prepareAssetsForSKRTests();
    }
    else {
      await prepareAssetsForOSSTests();
    }

    // Set port-forward to prometheus pod
    cancelPrometheusPortForward = prometheusPortForward();
  });

  after(async function() {
    // runs once after the last test in this block

    // Unregister SKR resources from Compass
    if (isSKR) {
      debug('Cleaning SKR...');
      await cleanCompassResourcesSKR(director, appName, scenarioName, skrInfo.compassID);
    }

    // Delete eventing backend secret if it was created by test
    if (eventMeshSecretFilePath !== "") {
      debug('Removing Event Mesh secret');
      await deleteEventingBackendK8sSecret(backendK8sSecretName, backendK8sSecretNamespace);
    }

    debug('Cleaning test resources');
    await cleanMockTestFixture(mockNamespace, testNamespace, true);

    cancelPrometheusPortForward();
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
    // Running Eventing tracing tests
    eventingTracingTestSuite();
    // Running Eventing Monitoring tests
    eventingMonitoringTest('nats');
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
    // Running Eventing Monitoring tests
    eventingMonitoringTest('beb');
  });

  context('with Nats backend switched back from BEB', function() {
    it("Switch Eventing Backend to Nats", async function () {
      await switchEventingBackend(backendK8sSecretName, backendK8sSecretNamespace, "nats");
      await waitForSubscriptionsTillReady(testNamespace)
    });

    // Running Eventing end-to-end tests
    eventingE2ETestSuite();
    // Running Eventing tracing tests
    eventingTracingTestSuite();
    // Running Eventing Monitoring tests
    eventingMonitoringTest('nats');
  });
});
