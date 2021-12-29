const axios = require("axios");
const https = require("https");
const httpsAgent = new https.Agent({
    rejectUnauthorized: false, // curl -k
});
axios.defaults.httpsAgent = httpsAgent;
const {
    appName,
    testNamespace,
    mockNamespace,
    isSKR,
    backendK8sSecretName,
    backendK8sSecretNamespace,
    eventMeshSecretFilePath,
    timeoutTime,
    slowTime,
    eventingScenarioName,
    skrInstanceId
} = require("./utils");
const {
    cleanMockTestFixture,
    cleanCompassResourcesSKR,
} = require("../test/fixtures/commerce-mock");
const {
    deleteEventingBackendK8sSecret,
} = require("../utils");
const {
    deprovisionSKR,
} = require("../kyma-environment-broker");
const {KEBClient, KEBConfig} = require("../kyma-environment-broker");

describe("Eventing tests cleanup", function () {
    this.timeout(timeoutTime);
    this.slow(slowTime);
    let director = null;
    let skrInfo = null;
    process.env.KEB_HOST = ""

    it("Cleaning: Test namespaces should be deleted", async function () {
        await cleanMockTestFixture(mockNamespace, testNamespace, true);

        // Delete eventing backend secret if it was created by test
        if (eventMeshSecretFilePath !== "") {
            await deleteEventingBackendK8sSecret(backendK8sSecretName, backendK8sSecretNamespace);
        }

        // Unregister SKR resources from Compass and deprovisions the cluster
        if (isSKR) {
            const keb = new KEBClient(KEBConfig.fromEnv());
            await deprovisionSKR(keb, skrInstanceId);
            await cleanCompassResourcesSKR(director, appName, eventingScenarioName, skrInfo.compassID);
        }
    });
});
