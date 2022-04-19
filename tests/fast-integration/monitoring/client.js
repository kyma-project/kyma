const axios = require('axios');
const https = require('https');

const {
  convertAxiosError,
  retryPromise,
  getVirtualService,
  debug,
  error,
} = require('../utils');

async function getGrafanaUrl() {
  const vs = await getVirtualService('kyma-system', 'monitoring-grafana');
  const host = vs.spec.hosts[0];

  return `https://${host}`;
}

async function getGrafanaDatasourceId(grafanaUrl, datasourceName) {
  const url = `${grafanaUrl}/api/datasources/id/${datasourceName}`;
  return retryPromise(async () => await axios.get(url), 5, 1000);
}

async function proxyGrafanaDatasource(datasourceName, path, retries, interval,
    timeout, debugMsg = undefined) {
  const grafanaUrl = await getGrafanaUrl();
  const datasourceResponse = await getGrafanaDatasourceId(grafanaUrl, datasourceName);
  const datasourceId = datasourceResponse.data.id;
  const url = `${grafanaUrl}/api/datasources/proxy/${datasourceId}/${path}`;
  debug('proxying Grafana datasource: ', url);
  return retryPromise(async () => {
    if (debugMsg) {
      debug(debugMsg);
    }
    return await axios.get(url, {timeout: timeout});
  }, retries, interval);
}

async function getPrometheusViaGrafana(path, retries = 5, interval = 30, timeout = 10000) {
  return await proxyGrafanaDatasource('Prometheus', path, retries, interval, timeout);
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

module.exports = {
  getGrafanaUrl,
  proxyGrafanaDatasource,
  getPrometheusActiveTargets,
  getPrometheusAlerts,
  getPrometheusRuleGroups,
  queryPrometheus,
  queryGrafana,
};
