const axios = require('axios');
const https = require('https');


const {
    kubectlPortForward,
    getResponse
} = require("../utils");

let prometheusPort = 9090;

function prometheusPortForward() {
    return kubectlPortForward("kyma-system", "prometheus-monitoring-prometheus-0", prometheusPort);
}

async function getPrometheusActiveTargets() {
    let path = "/api/v1/targets?state=active";
    let url = `http://localhost:${prometheusPort}${path}`;
    let responseBody = await getResponse(url, 30);
    return responseBody.data.data.activeTargets;
}

async function getPrometheusAlerts() {
    let path = "/api/v1/alerts";
    let url = `http://localhost:${prometheusPort}${path}`;
    let responseBody = await getResponse(url, 30);

    return responseBody.data.data.alerts;
}

async function getPrometheusRuleGroups() {
    let path = "/api/v1/rules";
    let url = `http://localhost:${prometheusPort}${path}`;
    let responseBody = await getResponse(url, 30);

    return responseBody.data.data.groups;
}

async function queryPrometheus(query) {
    let path = `/api/v1/query?query=${encodeURIComponent(query)}`;
    let url = `http://localhost:${prometheusPort}${path}`;
    let responseBody = await getResponse(url, 30);

    return responseBody.data.data.result;
}

async function queryGrafana(url, redirectURL, ignoreSSL, httpErrorCode) {
    try {
        // For more details see here: https://oauth2-proxy.github.io/oauth2-proxy/docs/behaviour
        delete axios.defaults.headers.common["Accept"]
        // Ignore SSL certificate for self signed certificates
        const agent = new https.Agent({
            rejectUnauthorized: !ignoreSSL
        });
        const res = await axios.get(url, { httpsAgent: agent })
        if (res.status === httpErrorCode) {
            if (res.request.res.responseUrl.includes(redirectURL)) {
                return true;
            }
        }
        return false;
    } catch(err) {
        const msg = "Error when querying Grafana: ";
        if (err.response) {
            if (err.response.status === httpErrorCode) {
                if (err.response.data.includes(redirectURL)) {
                    return true;
                }
            }
            console.log(msg + err.response.status + " : " + err.response.data)
            return false;
        } else {
            console.log(`${msg}: ${err.toString()}`);
            return false;
        }
    }
}

module.exports = {
    prometheusPortForward,
    getPrometheusActiveTargets,
    getPrometheusAlerts,
    getPrometheusRuleGroups,
    queryPrometheus,
    queryGrafana,
};
