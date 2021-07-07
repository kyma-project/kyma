const axios = require('axios');
const axiosRetry = require('axios-retry');

const {
    kubectlPortForward,
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
    axiosRetry(axios, {
        retries: 30,
        retryDelay: (retryCount) => {
            return retryCount * 5000;
        },
        retryCondition: (error) => {
            return !error.response || error.response.status != 200;
        },
    });

    let response = await axios.get(`http://localhost:${prometheusPort}${path}`, {
        timeout: 5000,
    });
    let responseBody = response.data;
    return responseBody;
}

async function queryGrafana(url, redirectURL, httpErrorCode) {
    try {
        delete axios.defaults.headers.common["Accept"]
        const res = await axios.get(url)
        if (res.status === httpErrorCode ) {
            if (res.request.res.responseUrl.includes(redirectURL)) {
                return true
            }
        }
        return false;
    } catch(err) {
        const msg = "Error when querying Grafana: ";
        if (err.response) {
            if (err.response.status === httpErrorCode) {
                if (err.response.data.includes(redirectURL)) {
                    return true
                }
            }
            console.log(msg + err.response.status + " : " + err.response.data)
            return false
        } else {
            console.log(`${msg}: ${err.toString()}`);
            return false
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
