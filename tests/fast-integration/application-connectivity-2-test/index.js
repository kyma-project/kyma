const installer = require('../installer')
const {
    ensureCommerceMockLocalTestFixture,
    checkAppGatewayResponse,
    sendEventAndCheckResponse,
    cleanMockTestFixture,
} = require("../test/fixtures/commerce-mock");
const {
    printRestartReport,
    getContainerRestartsForAllNamespaces,
} = require("../utils");

const kymaVersion = process.env.INSTALL_KYMA_VERSION || "1.20.0";


describe("Kyma Application Connectivity 2.0 tests", function () {

    this.timeout(10 * 60 * 1000);
    this.slow(5000);
    const testNamespace = "test"
    const skipComponents = ["dex","tracing","monitoring","console","kiali","logging"];
    const withCentralApplicationGateway = true;
    let initialRestarts = null;

    it(`Install Kyma ${kymaVersion}`, async function () {
        const resourcesPath = await installer.downloadCharts({ source: kymaVersion })
        await installer.installKyma({ resourcesPath, skipComponents, withCentralApplicationGateway })
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
