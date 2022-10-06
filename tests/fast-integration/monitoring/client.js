module.exports = {
  proxyGrafanaDatasource,
  getPrometheusActiveTargets,
  getPrometheusAlerts,
  getPrometheusRuleGroups,
  queryPrometheus,
  checkIfGrafanaIsReachable,
};

const axios = require('axios');
const https = require('https');
const httpsAgent = new https.Agent({
  rejectUnauthorized: false, // curl -k
});
axios.defaults.httpsAgent = httpsAgent;

const {
  convertAxiosError,
  retryPromise,
  getVirtualService,
  debug,
  error,
} = require('../utils');

async function proxyGrafanaDatasource(datasourceName, path, retries, interval,
    timeout, debugMsg = undefined) {
  const grafanaUrl = await getGrafanaUrl();
  const datasourceId = await getGrafanaDatasourceId(grafanaUrl, datasourceName);
  const url = `${grafanaUrl}/api/datasources/proxy/${datasourceId}/${path}`;

  return retryPromise(async () => {
    if (debugMsg) {
      debug(debugMsg);
    }
    debugMsg(`fetching grafana data source via: ${url}`);
    return await axios.get(url, {timeout: timeout});
  }, retries, interval);
}

async function getGrafanaUrl() {
  const vs = await getVirtualService('kyma-system', 'monitoring-grafana');
  const host = vs.spec.hosts[0];
  return `https://${host}`;
}

async function getGrafanaDatasourceId(grafanaUrl, datasourceName) {
  const url = `${grafanaUrl}/api/datasources/id/${datasourceName}`;
  const responseBody = await retryPromise(async () => await axios.get(url), 5, 1000);
  return responseBody.data.id;
}

async function getPrometheusActiveTargets() {
  const path = 'api/v1/targets?state=active';
  try {
    const responseBody = await getPrometheusViaGrafana(path);
    return responseBody.data.data.activeTargets;
  } catch (err) {
    throw convertAxiosError(err, 'cannot get prometheus targets');
  }
}

async function getPrometheusAlerts() {
  const path = 'api/v1/alerts';
  try {
    const responseBody = await getPrometheusViaGrafana(path);
    return responseBody.data.data.alerts;
  } catch (err) {
    throw convertAxiosError(err, 'cannot get prometheus alerts');
  }
}

async function getPrometheusRuleGroups() {
  const path = 'api/v1/rules';
  try {
    const responseBody = await getPrometheusViaGrafana(path);
    return responseBody.data.data.groups;
  } catch (err) {
    throw convertAxiosError(err, 'cannot get prometheus rules');
  }
}

async function queryPrometheus(query) {
  const path = `api/v1/query?query=${encodeURIComponent(query)}`;
  try {
    const responseBody = await getPrometheusViaGrafana(path);
    return responseBody.data.data.result;
  } catch (err) {
    throw convertAxiosError(err, 'cannot query prometheus');
  }
}

async function checkIfGrafanaIsReachable(redirectURL, httpErrorCode) {
  const url = await getGrafanaUrl();
  let ignoreSSL = false;
  if (url.includes('local.kyma.dev')) {
    ignoreSSL = true; // Ignore SSL certificate for self signed certificates
  }

  // For more details see here: https://oauth2-proxy.github.io/oauth2-proxy/docs/behaviour
  delete axios.defaults.headers.common['Accept'];
  const agent = new https.Agent({
    rejectUnauthorized: !ignoreSSL, // reject unauthorized when ssl should not be ignored
  });

  try {
    const response = await axios.get(url, {httpsAgent: agent});
    if (response.status === httpErrorCode && response.request.res.responseUrl.includes(redirectURL)) {
      return true;
    }
  } catch (err) {
    const msg = 'Error when querying Grafana: ';
    if (err.response) {
      if (err.response.status === httpErrorCode && err.response.data.includes(redirectURL)) {
        return true;
      }
    } else {
      error(`${msg}: ${err.toString()}`);
    }
  }

  return false;
}

async function getPrometheusViaGrafana(path, retries = 5, interval = 30, timeout = 10000) {
  return await proxyGrafanaDatasource('Prometheus', path, retries, interval, timeout);
}
