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
} = require("../test/fixtures/commerce-mock");
const {
  switchEventingBackend,
  waitForEventingBackendToReady,
} = require("../utils");

describe("Eventing tests with NATS", function () {
  this.timeout(10 * 60 * 1000);
  this.slow(5000);
  const testNamespace = "test";
  const testStartTimestamp = new Date().toISOString();

  it("CommerceMock test fixture should be ready", async function () {
    await ensureCommerceMockLocalTestFixture("mocks", testNamespace).catch((err) => {
      console.dir(err); // first error is logged
      return ensureCommerceMockLocalTestFixture("mocks", testNamespace);
    });
  });

  it("In-cluster event should be delivered (structured and binary mode)", async function () {
    await checkInClusterEventDelivery(testNamespace);
  });

  it("lastorder function should be reachable through secured API Rule", async function () {
    await checkFunctionResponse(testNamespace);
  });

  it("order.created.v1 event from CommerceMock should trigger the lastorder function", async function () {
    await sendEventAndCheckResponse();
  });

  it("Test namespaces should be deleted", async function () {
    await cleanMockTestFixture("mocks", testNamespace, true);
  });
});

describe("Eventing tests with BEB", function () {
  this.timeout(10 * 60 * 1000);
  this.slow(5000);
  const testNamespace = "test";
  const testStartTimestamp = new Date().toISOString();
  const eventMeshSecretName = process.env.BACKEND_SECRET_NAME || "eventing-backend";
  const eventMeshSecretNamespace = process.env.BACKEND_SECRET_NAMESPACE || "default";

  // runs once before the first test in this block
  before(async function() {
    console.log("switch eventing-backend to eventmesh")
    console.log(`eventMeshSecretName: ${eventMeshSecretName}, eventMeshSecretNamespace: ${eventMeshSecretNamespace}`)
    await switchEventingBackend(eventMeshSecretName, eventMeshSecretNamespace, "beb");
    await waitForEventingBackendToReady("eventing-backend", "kyma-system", "beb");
  });

  // runs once after the last test in this block
  after(async function() {
    console.log("switch eventing-backend to NATS")
    await switchEventingBackend(eventMeshSecretName, eventMeshSecretNamespace, "nats");
    await waitForEventingBackendToReady("eventing-backend", "kyma-system", "nats");
  });

  // tests
  it("CommerceMock test fixture should be ready", async function () {
    await ensureCommerceMockLocalTestFixture("mocks", testNamespace).catch((err) => {
      console.dir(err); // first error is logged
      return ensureCommerceMockLocalTestFixture("mocks", testNamespace);
    });
  });

  it("In-cluster event should be delivered (structured and binary mode)", async function () {
    await checkInClusterEventDelivery(testNamespace);
  });

  it("lastorder function should be reachable through secured API Rule", async function () {
    await checkFunctionResponse(testNamespace);
  });

  it("order.created.v1 event from CommerceMock should trigger the lastorder function", async function () {
    await sendEventAndCheckResponse();
  });

  it("Test namespaces should be deleted", async function () {
    await cleanMockTestFixture("mocks", testNamespace, true);
  });
});