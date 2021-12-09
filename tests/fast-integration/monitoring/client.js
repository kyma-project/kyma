const axios = require('axios');
const https = require('https');


const {
    kubectlPortForward,
    retryPromise,
    convertAxiosError,
    listPods,
    debug,
} = require("../utils");

const SECOND = 1000;
let prometheusPort = 9090;
let jaegerPort = 16686;

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

async function jaegerPortForward() {
    const res = await getJaegerPods();
    if (res.body.items.length === 0) {
        throw new Error("cannot find any jaeger pods");
    }

    return kubectlPortForward("kyma-system", res.body.items[0].metadata.name, jaegerPort);
}

async function getJaegerPods() {
    let labelSelector = `app=jaeger,app.kubernetes.io/component=all-in-one,app.kubernetes.io/instance=tracing-jaeger,app.kubernetes.io/managed-by=jaeger-operator,app.kubernetes.io/name=tracing-jaeger,app.kubernetes.io/part-of=jaeger`;
    return listPods(labelSelector, "kyma-system");
}

async function getJaegerTrace(traceId) {
    let path = `/api/traces/${traceId}`;
    let url = `http://localhost:${jaegerPort}${path}`;

    try {
        debug(`fetching trace: ${traceId} from jaeger`)
        let responseBody = await retryPromise(
            () => { 
                debug(`waiting for trace (id: ${traceId}) from jaeger...`)
                return axios.get(url, {timeout: 30 * SECOND})
            },
            30,
            1000
        );
        
        return responseBody.data; 
    } catch (err) {
        throw convertAxiosError(err, "cannot get jaeger trace");
    }
}

module.exports = {
    prometheusPortForward,
    getPrometheusActiveTargets,
    getPrometheusAlerts,
    getPrometheusRuleGroups,
    queryPrometheus,
    queryGrafana,
    jaegerPortForward,
    getJaegerTrace,
};
