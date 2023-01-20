const {
  debug,
  convertAxiosError,
  retryPromise,
  waitForVirtualService,
} = require('../utils');
const {proxyGrafanaDatasource} = require('../monitoring/client');

const axios = require('axios');
const https = require('https');
const httpsAgent = new https.Agent({
  rejectUnauthorized: false, // curl -k
});
axios.defaults.httpsAgent = httpsAgent;

async function getTestAppUrl() {
  const vs = await waitForVirtualService('default', 'svc-a');
  const host = vs.spec.hosts[0];
  return `https://${host}`;
}

async function callTestApp() {
  const testAppUrl = await getTestAppUrl();

  return retryPromise(async () => {
    return await axios.get(testAppUrl, {timeout: 10000});
  }, 5, 30);
}

async function getJaegerViaGrafana(path, retries = 5, interval = 30,
    timeout = 10000, debugMsg = undefined) {
  return await proxyGrafanaDatasource('Jaeger', path, retries, interval, timeout, debugMsg);
}

async function getJaegerTrace(traceId) {
  const path = `api/traces/${traceId}`;

  debug(`fetching trace: ${traceId} from jaeger`);

  try {
    const debugMsg = `waiting for trace (id: ${traceId}) from jaeger...`;
    const responseBody = await getJaegerViaGrafana(path, 30, 1000, 30 * 1000, debugMsg);
    return responseBody.data;
  } catch (err) {
    throw convertAxiosError(err, 'cannot get jaeger trace');
  }
}

async function getJaegerServices() {
  const path = `api/services`;

  debug(`fetching services from jaeger`);

  try {
    const debugMsg = `waiting for fetching service from jaeger...`;
    const responseBody = await getJaegerViaGrafana(path, 30, 1000, 30 * 1000, debugMsg);
    return responseBody.data;
  } catch (err) {
    throw convertAxiosError(err, 'cannot get jaeger trace');
  }
}

async function getJaegerTracesForService(serviceName, namespace = 'default') {
  const path = `api/traces?limit=20&lookback=2h&maxDuration&minDuration&service=${serviceName}.${namespace}`;

  debug(`fetching traces from jaeger`);

  try {
    const debugMsg = `waiting for fetching service from jaeger...`;
    const responseBody = await getJaegerViaGrafana(path, 30, 1000, 30 * 1000, debugMsg);
    return responseBody.data;
  } catch (err) {
    throw convertAxiosError(err, 'cannot get jaeger trace');
  }
}

module.exports = {
  getJaegerTrace,
  getJaegerServices,
  callTestApp,
  getJaegerTracesForService,
};
