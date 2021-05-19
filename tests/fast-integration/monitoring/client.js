const axios = require('axios');

const {
    kubectlPortForward,
} = require("../utils");

const {
    convertAxiosError,
    retryPromise,
} = require("../utils");

let prometheusPort = 9090;

function prometheusPortForward() {
    return kubectlPortForward("kyma-system", "prometheus-monitoring-prometheus-0", prometheusPort);
}

async function getPrometheusActiveTargets() {
    let responseBody = await get("/api/v1/targets?state=active");
    return responseBody.data.activeTargets;
}

async function getPrometheusAlerts() {
    let responseBody = await get("/api/v1/alerts");
    return responseBody.data.alerts;
}

async function getPrometheusRuleGroups() {
    let responseBody = await get("/api/v1/rules");
    return responseBody.data.groups;
}

async function queryPrometheus(query) {
    let responseBody = await get(`/api/v1/query?query=${encodeURIComponent(query)}`);
    return responseBody.data.result;
}

async function get(path) {
    let response = await axios.get(`http://localhost:${prometheusPort}${path}`);
    let responseBody = response.data;
    return responseBody;
}

module.exports = {
    prometheusPortForward,
    getPrometheusActiveTargets,
    getPrometheusAlerts,
    getPrometheusRuleGroups,
    queryPrometheus,
};
