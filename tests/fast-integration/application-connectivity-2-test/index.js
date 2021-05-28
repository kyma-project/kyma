const installer = require('../installer')
const axios = require("axios");
const https = require("https");
const { expect } = require("chai");
const httpsAgent = new https.Agent({
    rejectUnauthorized: false, // curl -k
});
axios.defaults.httpsAgent = httpsAgent;

const {
    ensureCommerceMockLocalTestFixture,
    checkAppGatewayResponse,
    sendEventAndCheckResponse,
    cleanMockTestFixture,
} = require("../test/fixtures/commerce-mock");
const {
    printRestartReport,
    getContainerRestartsForAllNamespaces,
    eventingSubscription,
    waitForFunction,
    waitForSubscription,
    k8sApply,
    genRandom,
    retryPromise,
    debug,
    waitForVirtualService,
} = require("../utils");


describe("Kyma Application Connectivity 2.0 tests", function () {

    this.timeout(10 * 60 * 1000);
    this.slow(5000);
    const testNamespace = "test"
    const skipComponents = ["dex","tracing","monitoring","console","kiali","logging"];
    const withCentralApplicationGateway = true;
    const newEventing = true;
    let initialRestarts = null;

    it("Install Kyma", async function() {
        await installer.installKyma({ skipComponents, newEventing, withCentralApplicationGateway });
    });

    it("Listing all pods in cluster", async function () {
        initialRestarts = await getContainerRestartsForAllNamespaces();
    });

    it("CommerceMock test fixture should be ready", async function () {
        await ensureCommerceMockLocalTestFixture("mocks", testNamespace, withCentralApplicationGateway).catch((err) => {
            console.dir(err); // first error is logged
            return ensureCommerceMockLocalTestFixture("mocks", testNamespace, withCentralApplicationGateway);
        });
    });

    it("lastorder function should be ready", async function () {
        await waitForFunction("lastorder", testNamespace);
    });

    it("In-cluster event subscription should be ready", async function () {
        await k8sApply([eventingSubscription(
            `sap.kyma.custom.inapp.order.received.v1`,
            `http://lastorder.${testNamespace}.svc.cluster.local`,
            "order-received",
            testNamespace)]);
        await waitForSubscription("order-received", testNamespace);
        await waitForSubscription("order-created", testNamespace);
    });

    it("in-cluster event should be delivered", async function () {
        const eventId = "event-"+genRandom(5);
        const vs = await waitForVirtualService(testNamespace, "lastorder");
        const mockHost = vs.spec.hosts[0];

        // send event using function query parameter send=true
        await retryPromise(() => axios.post(`https://${mockHost}`, { id: eventId }, {params:{send:true}}), 10, 1000)
        // verify if event was received using function query parameter inappevent=eventId
        await retryPromise(async () => {
            debug("Waiting for event: ",eventId);
            let response = await axios.get(`https://${mockHost}`, { params: { inappevent: eventId } })
            expect(response).to.have.nested.property("data.id", eventId, "The same event id expected in the result");
            expect(response).to.have.nested.property("data.shipped", true, "Order should have property shipped");
        }, 10, 1000);

    });

    it("function should reach Commerce mock API through app gateway", async function () {
        await checkAppGatewayResponse();
    });

    it("order.created.v1 event should trigger the lastorder function", async function () {
        await sendEventAndCheckResponse();
    });

    it("Should print report of restarted containers, skipped if no crashes happened", async function () {
        const afterTestRestarts = await getContainerRestartsForAllNamespaces();
        printRestartReport(initialRestarts, afterTestRestarts);
    });

    it("Test namespaces should be deleted", async function () {
        await cleanMockTestFixture("mocks", testNamespace, false);
    });
});
