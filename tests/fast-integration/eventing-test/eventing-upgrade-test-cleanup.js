const axios = require("axios");
const https = require("https");
const httpsAgent = new https.Agent({
    rejectUnauthorized: false, // curl -k
});
axios.defaults.httpsAgent = httpsAgent;
const {
    appName,
    scenarioName,
    testNamespace,
    mockNamespace,
    isSKR,
    backendK8sSecretName,
    backendK8sSecretNamespace,
    eventMeshSecretFilePath,
    timeoutTime,
    slowTime,
} = require("./utils");
const {
    cleanMockTestFixture,
    cleanCompassResourcesSKR,
} = require("../test/fixtures/commerce-mock");
const {
    deleteEventingBackendK8sSecret,
} = require("../utils");

describe("Eventing tests cleanup", function () {
    this.timeout(timeoutTime);
    this.slow(slowTime);
    let director = null;
    let skrInfo = null;


    it("Cleaning: Test namespaces should be deleted", async function () {
        await cleanMockTestFixture(mockNamespace, testNamespace, true);

        // Delete eventing backend secret if it was created by test
        if (eventMeshSecretFilePath !== "") {
            await deleteEventingBackendK8sSecret(backendK8sSecretName, backendK8sSecretNamespace);
        }

        // Unregister SKR resources from Compass
        if (isSKR) {
            await cleanCompassResourcesSKR(director, appName, scenarioName, skrInfo.compassID);
        }
    });
});
