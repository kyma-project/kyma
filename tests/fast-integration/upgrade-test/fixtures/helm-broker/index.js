const k8s = require("@kubernetes/client-node");
const fs = require("fs");
const path = require("path");
const { expect } = require("chai");
const https = require("https");
const axios = require("axios").default;
const httpsAgent = new https.Agent({
    rejectUnauthorized: false, // curl -k
});
axios.defaults.httpsAgent = httpsAgent;

const {
    k8sApply,
    waitForServiceClass,
    waitForServiceInstance
} = require("../../../utils");

const sampleAddonsYaml = fs.readFileSync(
    path.join(__dirname, "./sample-addons.yaml"),
    {
        encoding: "utf8",
    }
);

const testingServiceInstanceYaml = fs.readFileSync(
    path.join(__dirname, "./service-instance.yaml"),
    {
        encoding: "utf8",
    }
);

const clusterAddonsCfgObj = k8s.loadYaml(sampleAddonsYaml);
const serviceInstanceObj = k8s.loadYaml(testingServiceInstanceYaml);

async function ensureHelmBrokerTestFixture(targetNamespace) {
    await k8sApply(clusterAddonsCfgObj);
    await waitForServiceClass("testing", targetNamespace)
    await k8sApply(serviceInstanceObj, );
    await waitForServiceInstance('testing', targetNamespace);
}

module.exports = {
    ensureHelmBrokerTestFixture,
};