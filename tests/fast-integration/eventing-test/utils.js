const {
  cleanMockTestFixture,
  cleanCompassResourcesSKR,
} = require('../test/fixtures/commerce-mock');

const {
  debug,
  getEnvOrThrow,
  deleteEventingBackendK8sSecret,
  deleteK8sConfigMap,
  getShootNameFromK8sServerUrl,
  listPods,
  retryPromise,
} = require('../utils');

const {DirectorClient, DirectorConfig, getAlreadyAssignedScenarios} = require('../compass');
const {GardenerClient, GardenerConfig} = require('../gardener');
const {eventMeshSecretFilePath} = require('./common/common');
const axios = require('axios');
const kymaVersion = process.env.KYMA_VERSION || '';
const isSKR = process.env.KYMA_TYPE === 'SKR';
const skrInstanceId = process.env.INSTANCE_ID || '';
const testCompassFlow = process.env.TEST_COMPASS_FLOW || false;
const skipResourceCleanup = process.env.SKIP_CLEANUP || false;
const suffix = getSuffix(isSKR, testCompassFlow);
const appName = `app-${suffix}`;
const scenarioName = `test-${suffix}`;
const testNamespace = `test-${suffix}`;
const mockNamespace = process.env.MOCK_NAMESPACE || 'mocks';
const backendK8sSecretName = process.env.BACKEND_SECRET_NAME || 'eventing-backend';
const backendK8sSecretNamespace = process.env.BACKEND_SECRET_NAMESPACE || 'default';
const streamDataConfigMapName = 'eventing-stream-info';
const eventingNatsSvcName = 'eventing-nats';
const eventingNatsApiRuleAName = `${eventingNatsSvcName}-apirule`;
const timeoutTime = 10 * 60 * 1000;
const slowTime = 5000;
const streamConfig = { };
const subscriptionNames = {
  orderCreated: 'order-created',
  orderReceived: 'order-received',
};

// SKR related constants
let gardener = null;
let director = null;
let shootName = null;
if (isSKR) {
  gardener = new GardenerClient(GardenerConfig.fromEnv()); // create gardener client
  shootName = getShootNameFromK8sServerUrl();

  if (testCompassFlow) {
    director = new DirectorClient(DirectorConfig.fromEnv()); // director client for Compass
  }
}

// cleans up all the test resources including the compass scenario
async function cleanupTestingResources() {
  if (isSKR && testCompassFlow) {
    debug('Cleaning compass resources');
    // Get shoot info from gardener to get compassID for this shoot
    const skrInfo = await gardener.getShoot(shootName);
    await cleanCompassResourcesSKR(director, appName, scenarioName, skrInfo.compassID);
  }

  // skip the cluster resources cleanup if the SKIP_CLEANUP env flag is set
  if (skipResourceCleanup === 'true') {
    return;
  }

  // Delete eventing backend secret if it was created by test
  if (eventMeshSecretFilePath) {
    debug('Removing Event Mesh secret');
    await deleteEventingBackendK8sSecret(backendK8sSecretName, backendK8sSecretNamespace);
  }

  debug('Removing JetStream data configmap');
  await deleteK8sConfigMap(streamDataConfigMapName);

  debug(`Removing ${testNamespace} and ${mockNamespace} namespaces`);
  await cleanMockTestFixture(mockNamespace, testNamespace, true);
}

// gets the suffix depending on kyma type
function getSuffix(isSKR, testCompassFlow) {
  let suffix;
  if (isSKR && testCompassFlow) {
    suffix = getEnvOrThrow('TEST_SUFFIX');
  } else {
    suffix = 'evnt';
  }
  return suffix;
}

// getRegisteredCompassScenarios lists the registered compass scenarios
async function getRegisteredCompassScenarios() {
  try {
    const skrInfo = await gardener.getShoot(shootName);
    const result = await getAlreadyAssignedScenarios(director, skrInfo.compassID);
    console.log('List of the active scenarios:');
    result.map((scenario, i) => console.log('%s)%s', i+1, scenario));
  } catch (e) {
    console.log('Cannot display the assigned scenarios');
  }
}

async function getNatsPods() {
  const labelSelector = 'app.kubernetes.io/name=nats';
  return await listPods(labelSelector, 'kyma-system');
}

// getStreamConfigForJetStream gets the stream retention policy and the consumer deliver policy env variables
// from the eventing controller pod and also checks if these env variables exist on the pod.
async function getStreamConfigForJetStream() {
  const labelSelector = 'app.kubernetes.io/instance=eventing,app.kubernetes.io/name=controller';
  const res = await listPods(labelSelector, 'kyma-system');
  let envsCount = 0;
  res.body?.items[0]?.spec.containers.find((container) =>
    container.name === 'controller',
  ).env.forEach((env) => {
    if (env.name === 'JS_STREAM_RETENTION_POLICY') {
      streamConfig['retention_policy'] = env.value;
      envsCount++;
    }
    if (env.name === 'JS_CONSUMER_DELIVER_POLICY') {
      streamConfig['consumer_deliver_policy'] = env.value;
      envsCount++;
    }
  });
  // check to make sure the environment variables exist
  return envsCount === 2;
}

async function getJetStreamStreamData(host) {
  const responseJson = await retryPromise(async () => await axios.get(`https://${host}/jsz?streams=true`), 5, 1000);
  const streamName = responseJson.data.account_details[0].stream_detail[0].name;
  const streamCreationTime = responseJson.data.account_details[0].stream_detail[0].created;

  return {
    streamName: streamName,
    streamCreationTime: streamCreationTime,
  };
}

function skipAtLeastOnceDeliveryTest() {
  return !(streamConfig['retention_policy'] === 'limits' &&
      streamConfig['consumer_deliver_policy'] === 'all');
}

function skipStreamReCreationTest(streamData) {
  return streamData.streamCreationTime === undefined;
}

module.exports = {
  appName,
  scenarioName,
  testNamespace,
  mockNamespace,
  kymaVersion,
  isSKR,
  skrInstanceId,
  testCompassFlow,
  backendK8sSecretName,
  backendK8sSecretNamespace,
  streamDataConfigMapName,
  eventingNatsSvcName,
  eventingNatsApiRuleAName,
  timeoutTime,
  slowTime,
  director,
  gardener,
  shootName,
  suffix,
  cleanupTestingResources,
  getRegisteredCompassScenarios,
  getNatsPods,
  getStreamConfigForJetStream,
  skipAtLeastOnceDeliveryTest,
  getJetStreamStreamData,
  skipStreamReCreationTest,
  subscriptionNames,
};
