const axios = require('axios');
const https = require('https');


const {
    kubectlPortForward,
    retryPromise,
    convertAxiosError,
} = require("../utils");

let prometheusPort = 9090;

function prometheusPortForward() {
    return kubectlPortForward("kyma-system", "prometheus-monitoring-prometheus-0", prometheusPort);
}

async function getPrometheusActiveTargets() {
    let path = "/api/v1/targets?state=active";
    let url = `http://localhost:${prometheusPort}${path}`;
    try {
        let responseBody = await retryPromise(() => axios.get(url, {timeout: 10000}), 5);
        return responseBody.data.data.activeTargets;
    } catch(err) {
        throw convertAxiosError(err, "cannot get prometheus targets");
    }
}

async function getPrometheusAlerts() {
    let path = "/api/v1/alerts";
    let url = `http://localhost:${prometheusPort}${path}`;
    try {
        let responseBody = await retryPromise(() => axios.get(url, {timeout: 10000}), 5);
        return responseBody.data.data.alerts; 
    } catch (err) {
        throw convertAxiosError(err, "cannot get prometheus alerts");
    }
}

async function getPrometheusRuleGroups() {
    let path = "/api/v1/rules";
    let url = `http://localhost:${prometheusPort}${path}`;
    try {
        let responseBody = await retryPromise(() => axios.get(url, {timeout: 10000}), 5);
        return responseBody.data.data.groups;
    } catch (err) {
        throw convertAxiosError(err, "cannot get prometheus rules");
    }
}

async function queryPrometheus(query) {
    let path = `/api/v1/query?query=${encodeURIComponent(query)}`;
    let url = `http://localhost:${prometheusPort}${path}`;
    try {
        let responseBody = await retryPromise(() => axios.get(url, {timeout: 10000}), 5);
        return responseBody.data.data.result;
    } catch (err) {
        throw convertAxiosError(err, "cannot query prometheus");
    }
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
