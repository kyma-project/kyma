const axios = require('axios');
const https = require('https');

const {
  kc,
  kubectlPortForward,
  retryPromise,
  convertAxiosError,
  listPods,
  debug,
} = require('../utils');

const SECOND = 1000;
const jaegerPort = 16686;

function prometheusGet(path) {
  let httpsAgent;
  let headers;
  const token = kc.getCurrentUser().token;
  const caCrt = Buffer.from(kc.getCurrentCluster().caData, 'base64').toString();
  if (token) {
    httpsAgent = new https.Agent({
      rejectUnauthorized: false,
      ca: caCrt,
      timeout: 10000,
    });
    headers = {'Authorization': `Bearer ${token}`};
  }

  const server = kc.getCurrentCluster().server;
  const prometheusProxyUrl = 'api/v1/namespaces/kyma-system/services/monitoring-prometheus:http-web/proxy';
  const url = `${server}/${prometheusProxyUrl}${path}`;

  return retryPromise(() => axios.get(url, {httpsAgent: httpsAgent, headers: headers}), 5);
}

async function getPrometheusActiveTargets() {
  const path = '/api/v1/targets?state=active';
  try {
    const responseBody = await prometheusGet(path);
    return responseBody.data.data.activeTargets;
  } catch (err) {
    throw convertAxiosError(err, 'cannot get prometheus targets');
  }
}

async function getPrometheusAlerts() {
  const path = '/api/v1/alerts';
  try {
    const responseBody = await prometheusGet(path);
    return responseBody.data.data.alerts;
  } catch (err) {
    throw convertAxiosError(err, 'cannot get prometheus alerts');
  }
}

async function getPrometheusRuleGroups() {
  const path = '/api/v1/rules';
  try {
    const responseBody = await prometheusGet(path);
    return responseBody.data.data.groups;
  } catch (err) {
    throw convertAxiosError(err, 'cannot get prometheus rules');
  }
}

async function queryPrometheus(query) {
  const path = `/api/v1/query?query=${encodeURIComponent(query)}`;
  try {
    const responseBody = await prometheusGet(path);
    return responseBody.data.data.result;
  } catch (err) {
    throw convertAxiosError(err, 'cannot query prometheus');
  }
}

async function queryGrafana(url, redirectURL, ignoreSSL, httpErrorCode) {
  try {
    // For more details see here: https://oauth2-proxy.github.io/oauth2-proxy/docs/behaviour
    delete axios.defaults.headers.common['Accept'];
    // Ignore SSL certificate for self signed certificates
    const agent = new https.Agent({
      rejectUnauthorized: !ignoreSSL,
    });
    const res = await axios.get(url, {httpsAgent: agent});
    if (res.status === httpErrorCode) {
      if (res.request.res.responseUrl.includes(redirectURL)) {
        return true;
      }
    }
    return false;
  } catch (err) {
    const msg = 'Error when querying Grafana: ';
    if (err.response) {
      if (err.response.status === httpErrorCode) {
        if (err.response.data.includes(redirectURL)) {
          return true;
        }
      }
      console.log(msg + err.response.status + ' : ' + err.response.data);
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
    throw new Error('cannot find any jaeger pods');
  }

  return kubectlPortForward('kyma-system', res.body.items[0].metadata.name, jaegerPort);
}

async function getJaegerPods() {
  const labelSelector = 'app=jaeger,' +
    'app.kubernetes.io/component=all-in-one,' +
    'app.kubernetes.io/instance=tracing-jaeger,' +
    'app.kubernetes.io/managed-by=jaeger-operator,' +
    'app.kubernetes.io/name=tracing-jaeger,' +
    'app.kubernetes.io/part-of=jaeger';
  return listPods(labelSelector, 'kyma-system');
}

async function getJaegerTrace(traceId) {
  const path = `/api/traces/${traceId}`;
  const url = `http://localhost:${jaegerPort}${path}`;

  try {
    debug(`fetching trace: ${traceId} from jaeger`);
    const responseBody = await retryPromise(
        () => {
          debug(`waiting for trace (id: ${traceId}) from jaeger...`);
          return axios.get(url, {timeout: 30 * SECOND});
        },
        30,
        1000,
    );

    return responseBody.data;
  } catch (err) {
    throw convertAxiosError(err, 'cannot get jaeger trace');
  }
}

module.exports = {
  getPrometheusActiveTargets,
  getPrometheusAlerts,
  getPrometheusRuleGroups,
  queryPrometheus,
  queryGrafana,
  jaegerPortForward,
  getJaegerTrace,
};
