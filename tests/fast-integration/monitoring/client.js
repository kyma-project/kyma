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

async function queryGrafana(url) {
    console.log("ggg: " + url)

    try {
        console.log("ggg1")

        const responseBody = await axios.get(url);
        console.log("foo: ", responseBody)
        return responseBody.data;
    } catch(err) {
        console.log("ggg2")

        const msg = "Error when querying Grafana";
        if (err.response) {
            console.log("ggg3: " + err.response.statusText)
        

            throw new Error(`${msg}: ${err.response.status} ${err.response.statusText}: ${err.response.data}`);
        } else {
            console.log("ggg4")

            throw new Error(`${msg}: ${err.toString()}`);
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
