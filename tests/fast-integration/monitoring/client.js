const axios = require('axios');
const https = require('https');

const {
  callServiceViaProxy,
  convertAxiosError,
  retryPromise,
  getVirtualService,
  debug,
  error, info,
} = require('../utils');

function getPrometheus(path) {
  return callServiceViaProxy('kyma-system', 'monitoring-prometheus', '9090', path);
}

async function getJaegerViaGrafana(path, retries, interval, timeout, debugMsg) {
  const grafanaUrl = await getGrafanaUrl();
  const url = `${grafanaUrl}/api/datasources/proxy/3/jaeger/${path}`;
  info('jaeger grafana url', url);

  delete axios.defaults.headers.common['Accept'];
  const httpsAgent = new https.Agent({
    rejectUnauthorized: false,
  });
  return retryPromise(async () => {
    if (debugMsg) {
      debug(debugMsg);
    }
    return await axios.get(url, {httpsAgent: httpsAgent, timeout: timeout,
      method: opts.method, headers: opts.headers, data: opts.data});
  }, retries, interval);
}

async function getPrometheusActiveTargets() {
  const path = 'api/v1/targets?state=active';
  try {
    const responseBody = await getPrometheus(path);
    return responseBody.data.data.activeTargets;
  } catch (err) {
    throw convertAxiosError(err, 'cannot get prometheus targets');
  }
}

async function getPrometheusAlerts() {
  const path = 'api/v1/alerts';
  try {
    const responseBody = await getPrometheus(path);
    return responseBody.data.data.alerts;
  } catch (err) {
    throw convertAxiosError(err, 'cannot get prometheus alerts');
  }
}

async function getPrometheusRuleGroups() {
  const path = 'api/v1/rules';
  try {
    const responseBody = await getPrometheus(path);
    return responseBody.data.data.groups;
  } catch (err) {
    throw convertAxiosError(err, 'cannot get prometheus rules');
  }
}

async function queryPrometheus(query) {
  const path = `api/v1/query?query=${encodeURIComponent(query)}`;
  try {
    const responseBody = await getPrometheus(path);
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
      error(msg + err.response.status + ' : ' + err.response.data);
      return false;
    } else {
      error(`${msg}: ${err.toString()}`);
      return false;
    }
  }
}

async function getJaegerTrace(traceId) {
  const path = `api/traces/${traceId}`;

  const retries = 30;
  const interval = 1000;
  const timeout = 30 * 1000;
  const debugMsg = `waiting for trace (id: ${traceId}) from jaeger...`;
  debug(`fetching trace: ${traceId} from jaeger`);

  try {
    const response = getJaegerViaGrafana(path, retries, interval, timeout, debugMsg);
    info('jaeger.response.data', response.data);
    return response.data;
  } catch (err) {
    throw convertAxiosError(err, 'cannot get jaeger trace');
  }
}

async function getGrafanaUrl() {
  const vs = await getVirtualService('kyma-system', 'monitoring-grafana');
  const hosts = vs.spec.hosts[0];

  return `https://${hosts}`;
}

module.exports = {
  getPrometheusActiveTargets,
  getPrometheusAlerts,
  getPrometheusRuleGroups,
  queryPrometheus,
  queryGrafana,
  getGrafanaUrl,
  getJaegerTrace,
};
