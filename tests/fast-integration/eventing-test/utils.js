const {
  cleanMockTestFixture,
  generateTraceParentHeader,
  checkTrace,
} = require('../test/fixtures/commerce-mock');

const {
  debug,
  getEnvOrThrow,
  deleteEventingBackendK8sSecret,
  deleteK8sConfigMap,
  getShootNameFromK8sServerUrl,
  retryPromise,
  waitForVirtualService,
  k8sApply,
  k8sDelete,
  waitForFunction,
  eventingSubscription,
  waitForSubscription,
  eventingSubscriptionV1Alpha2,
  convertAxiosError,
  sleep,
  getConfigMap,
  createK8sConfigMap,
  namespaceObj,
} = require('../utils');

const {DirectorClient, DirectorConfig, getAlreadyAssignedScenarios} = require('../compass');
const {GardenerClient, GardenerConfig} = require('../gardener');
const {eventMeshSecretFilePath, getEventMeshNamespace} = require('./common/common');
const axios = require('axios');
const {v4: uuidv4} = require('uuid');
const fs = require('fs');
const path = require('path');
const k8s = require('@kubernetes/client-node');
const {expect} = require('chai');

// Variables
const kymaVersion = process.env.KYMA_VERSION || '';
const kymaStreamName = 'sap';
const isSKR = process.env.KYMA_TYPE === 'SKR';
const skrInstanceId = process.env.INSTANCE_ID || '';
const testCompassFlow = process.env.TEST_COMPASS_FLOW === 'true';
const isUpgradeJob = process.env.EVENTING_UPGRADE_JOB === 'true';
const isJSRecreatedTestEnabled = process.env.EVENTING_JS_RECREATED_TEST === 'true';
const isJSAtLeastOnceDeliveryTestEnabled = process.env.EVENTING_JS_ATLEASTONCE_TEST === 'true';
const skipResourceCleanup = process.env.SKIP_CLEANUP || false;
const suffix = getSuffix(isSKR, testCompassFlow);
const appName = `app-${suffix}`;
const testNamespace = `test-${suffix}`;
const backendK8sSecretName = process.env.BACKEND_SECRET_NAME || 'eventing-backend';
const backendK8sSecretNamespace = process.env.BACKEND_SECRET_NAMESPACE || 'default';
const testDataConfigMapName = 'eventing-test-data';
const jsRecreatedTestConfigMapName = 'eventing-fi-js-recreated-test';
const eventingNatsSvcName = 'eventing-nats';
const eventingNatsApiRuleAName = `${eventingNatsSvcName}-apirule`;
const timeoutTime = 10 * 60 * 1000;
const slowTime = 5000;
const eppInClusterUrl = 'eventing-event-publisher-proxy.kyma-system';
const eventingSinkName = 'eventing-sink';
const eventingUpgradeSinkName = 'eventing-upgrade-sink';

// ****** Event types to test ***********//
const v1alpha1SubscriptionsTypes = [
  {
    name: 'fi-test-sub-0',
    type: 'sap.kyma.custom.noapp.order.tested.v1',
  },
  {
    name: 'fi-test-sub-1',
    type: 'sap.kyma.custom.connected-app.order.tested.v1',
  },
  {
    name: 'fi-test-sub-2',
    type: 'sap.kyma.custom.test-app.order-$.second.R-e-c-e-i-v-e-d.v1',
  },
  {
    name: 'fi-test-sub-3',
    type: 'sap.kyma.custom.connected-app2.or-der.crea-ted.one.two.three.v4',
  },
];

const subscriptionsTypes = [
  {
    name: 'fi-test-sub-v2-0',
    type: 'order.modified.v1',
    source: 'myapp',
    consumerName: 'e04ea2aff4332541145342207495afce',
  },
  {
    name: 'fi-test-sub-v2-1',
    type: 'or-der.crea-ted.one.two.three.four.v4',
    source: 'test-app',
  },
  {
    name: 'fi-test-sub-v2-2',
    type: 'Order-$.third.R-e-c-e-i-v-e-d.v1',
    source: 'test-app',
  },
];

const subscriptionsExactTypeMatching = [
  {
    name: 'fi-test-sub-v2-exact-0',
    type: 'sap.kyma.custom.exact.order.completed.v2',
    source: undefined,
    typeMatching: 'exact',
  },
];

// ****** ************* ***********//

// SKR related constants
let gardener = null;
let director = null;
let shootName = null;
if (isSKR && skrInstanceId && skrInstanceId !== '') {
  gardener = new GardenerClient(GardenerConfig.fromEnv()); // create gardener client
  shootName = getShootNameFromK8sServerUrl();

  if (testCompassFlow) {
    director = new DirectorClient(DirectorConfig.fromEnv()); // director client for Compass
  }
}

// cleans up all the test resources including the compass scenario
async function cleanupTestingResources() {
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
  await deleteK8sConfigMap(testDataConfigMapName);
  await deleteK8sConfigMap(jsRecreatedTestConfigMapName);

  debug(`Removing ${testNamespace} and mocks namespaces`);
  await cleanMockTestFixture('mocks', testNamespace, true);
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

async function getClusterHost(apiRuleName, namespace) {
  const vs = await waitForVirtualService(namespace, apiRuleName);
  const mockHost = vs.spec.hosts[0];
  return mockHost.split('.').slice(1).join('.');
}

function createNewEventId() {
  return uuidv4();
}

function getK8sFunctionObject(funcName) {
  const functionYaml = fs.readFileSync(
      path.join(__dirname, `./assets/${funcName}.yaml`),
      {
        encoding: 'utf8',
      },
  );

  return k8s.loadAllYaml(functionYaml);
}

async function k8sApplyWithRetries(resources, namespace, patch = true, retries = 5, interval=1500) {
  return retryPromise(async () => await k8sApply(resources, namespace, patch), retries, interval);
}

async function k8sDeleteWithRetries(listOfSpecs, namespace, retries = 5, interval=1500) {
  return retryPromise(async () => await k8sDelete(listOfSpecs, namespace), retries, interval);
}

async function deployEventingSinkFunction(funcName) {
  await k8sApplyWithRetries(getK8sFunctionObject(funcName), testNamespace, true);
}

async function undeployEventingFunction(funcName) {
  await k8sDeleteWithRetries(getK8sFunctionObject(funcName), testNamespace);
}

async function waitForEventingSinkFunction(funcName) {
  await waitForFunction(funcName, testNamespace, 300000);
}

async function deployV1Alpha1Subscriptions() {
  const sink = `http://${eventingSinkName}.${testNamespace}.svc.cluster.local`;
  debug(`Using sink: ${sink}`);

  // creating v1alpha1 subscriptions
  for (let i=0; i < v1alpha1SubscriptionsTypes.length; i++) {
    const subName = v1alpha1SubscriptionsTypes[i].name;
    const eventType = v1alpha1SubscriptionsTypes[i].type;

    debug(`Creating subscription: ${subName} with type: ${eventType}`);
    await k8sApplyWithRetries([eventingSubscription(eventType, sink, subName, testNamespace)]);
    debug(`Waiting for subscription: ${subName} with type: ${eventType}`);
    await waitForSubscription(subName, testNamespace);
  }
}

async function deployV1Alpha2Subscriptions() {
  const sink = `http://${eventingSinkName}.${testNamespace}.svc.cluster.local`;
  debug(`Using sink: ${sink}`);

  // creating v1alpha2 subscriptions - standard type matching
  for (let i=0; i < subscriptionsTypes.length; i++) {
    const subName = subscriptionsTypes[i].name;
    const eventType = subscriptionsTypes[i].type;
    const eventSource = subscriptionsTypes[i].source;

    debug(`Creating subscription: ${subName} with type: ${eventType}, source: ${eventSource}`);
    await k8sApplyWithRetries([eventingSubscriptionV1Alpha2(eventType, eventSource, sink, subName, testNamespace)]);
    debug(`Waiting for subscription: ${subName} with type: ${eventType}, source: ${eventSource}`);
    await waitForSubscription(subName, testNamespace);
  }

  // creating v1alpha2 subscriptions - exact type matching
  for (let i=0; i < subscriptionsExactTypeMatching.length; i++) {
    const subName = subscriptionsExactTypeMatching[i].name;
    const eventType = subscriptionsExactTypeMatching[i].type;
    let eventSource = subscriptionsExactTypeMatching[i].source;
    if (!subscriptionsTypes[i].source) {
      eventSource = getEventMeshNamespace();
    }

    debug(`Creating subscription (TypeMatching: exact): ${subName} with type: ${eventType}, source: ${eventSource}`);
    await k8sApplyWithRetries(
        [eventingSubscriptionV1Alpha2(eventType, eventSource, sink, subName, testNamespace, 'exact')]);
    debug(`Waiting for subscription: ${subName} with type: ${eventType}, source: ${eventSource}`);
    await waitForSubscription(subName, testNamespace);
  }
}

async function waitForV1Alpha1Subscriptions() {
  // waiting for v1alpha1 subscriptions
  for (let i=0; i < v1alpha1SubscriptionsTypes.length; i++) {
    const subName = v1alpha1SubscriptionsTypes[i].name;
    debug(`Waiting for subscription: ${subName} with type: ${v1alpha1SubscriptionsTypes[i].type}`);
    await waitForSubscription(subName, testNamespace);
  }
}

async function waitForV1Alpha2Subscriptions() {
  // waiting for v1alpha2 subscriptions
  for (let i=0; i < subscriptionsTypes.length; i++) {
    const subName = subscriptionsTypes[i].name;
    debug(`Waiting for subscription: ${subName} with type: ${subscriptionsTypes[i].type}`);
    await waitForSubscription(subName, testNamespace);
  }

  // waiting for v1alpha2 subscriptions - exact type matching
  for (let i=0; i < subscriptionsExactTypeMatching.length; i++) {
    const subName = subscriptionsExactTypeMatching[i].name;
    debug(`Waiting for subscription: ${subName} with type: ${subscriptionsExactTypeMatching[i].type}`);
    await waitForSubscription(subName, testNamespace);
  }
}

async function checkFunctionReachable(name, namespace, host) {
  // the function should be reachable.
  const res = await retryPromise(
      () => axios.post(`https://${name}.${host}/function`, {orderCode: '789'}, {
        timeout: 5000,
      }),
      45,
      2000,
  ).catch((err) => {
    debug(`Error when trying to reach the function: ${name}`);
    debug(err);
    throw convertAxiosError(err, `Function ${name} responded with error`);
  });

  // the request should be authorized and successful
  expect(res.status).to.be.equal(200);
}

async function checkFunctionUnreachable(name, namespace, host) {
  return await retryPromise(
      async () => {
        const response = await axios.post(`https://${name}.${host}/function`, {orderCode: '789'}, {
          timeout: 5000,
        });
        expect(response.status).to.not.equal(200);
        return response;
      }, 45, 2 * 1000)
      .catch((err) => {
        debug(err);
      });
}

async function checkEventTracing(proxyHost, eventType, eventSource, namespace) {
  // first send an event and verify if it was delivered
  const result = await checkEventDelivery(proxyHost, 'binary', eventType, eventSource);
  expect(result).to.have.nested.property('traceParentId');
  expect(result.traceParentId).to.not.be.empty;
  expect(result.response).to.have.nested.property('data.metadata.podName');
  expect(result.response.data.metadata.podName).to.not.be.empty;

  // Define expected trace data
  const podName = result.response.data.metadata.podName;
  const correctTraceProcessSequence = [
    // We are sending the in-cluster event from inside the eventing sink pod
    'istio-ingressgateway.istio-system',
    `${eventingSinkName}-${podName.split('-')[2]}.${namespace}`,
    'eventing-publisher-proxy.kyma-system',
    'eventing-controller.kyma-system',
    `${eventingSinkName}-${podName.split('-')[2]}.${namespace}`,
  ];

  // wait sometime for jaeger to complete tracing data.
  // Arrival of traces might be delayed by otel-collectors batch timeout.
  const traceId = result.traceParentId.split('-')[1];
  debug(`Checking the tracing with traceId: ${traceId}, traceParentId: ${result.traceParentId}`);
  await sleep(20_000);
  await checkTrace(traceId, correctTraceProcessSequence);
}

// checks if the event publish and receive is working.
// Possible values for encoding are [binary, structured, legacy].
async function checkEventDelivery(proxyHost, encoding, eventType, eventSource,
    isSubV1Alpha1 = false, sinkName=eventingSinkName) {
  const eventId = createNewEventId();

  debug(`Publishing event with id:${eventId}, type: ${eventType}, source: ${eventSource}...`);
  const result = await publishEventWithRetry(proxyHost, encoding, eventId, eventType, eventSource, isSubV1Alpha1);

  debug(`Verifying if event with id:${eventId}, type: ${eventType}, source: ${eventSource} was received by sink...`);
  const result2 = await ensureEventReceivedWithRetry(sinkName, proxyHost,
      encoding, eventId, eventType, eventSource);
  return {
    eventId,
    traceParentId: result.traceParentId,
    response: result2.response,
  };
}

// send event using function query parameter send=true
async function publishEventWithRetry(proxyHost, encoding, eventId, eventType, eventSource,
    isSubV1Alpha1 = false, retriesLeft = 10) {
  return retryPromise(async () => {
    let reqBody = {};
    const traceParentId = await generateTraceParentHeader();

    if (encoding === 'binary') { // binary CE
      reqBody = createBinaryCloudEventRequestBody(eventId, eventType, eventSource, traceParentId);
    } else if (encoding === 'structured') { // structured CE
      reqBody = createStructuredCloudEventRequestBody(eventId, eventType, eventSource, traceParentId);
    } else if (encoding === 'legacy') {
      reqBody = createLegacyEventRequestBody(eventId, eventType, eventSource, isSubV1Alpha1);
    } else {
      throw new Error('Invalid encoding. Possible values are [binary, structured, legacy]');
    }

    // console out information
    debug(`Sending Event request to ${eventingSinkName}:`, reqBody);

    // send request
    const response = await axios.post(`https://${eventingSinkName}.${proxyHost}`, reqBody, {
      params: {
        send: true,
      },
      headers: {
        'traceparent': traceParentId,
      },
    });

    debug(`Response from ${eventingSinkName} for sending event:`, {
      status: response.status,
      data: response.data,
    });

    if (response.data.success !== true) {
      throw convertAxiosError(response.data.errorMessage);
    }
    expect(response.status).to.be.equal(200);
    // EPP response should be 204 (or 200 for legacy)
    expect(response.data.status).to.be.oneOf([200, 204]);

    return {
      traceParentId,
      response,
    };
  }, retriesLeft, 1000);
}

// verify if event was received using function
async function ensureEventReceivedWithRetry(sink, proxyHost,
    encoding, eventId, eventType, eventSource, retriesLeft = 10) {
  return await retryPromise(async () => {
    debug(`Waiting to receive CE event "${eventId}"`);

    const response = await axios.get(`https://${sink}.${proxyHost}`,
        {params: {eventid: eventId}});

    debug('Received response:', {
      status: response.status,
      statusText: response.statusText,
      data: response.data,
    });

    if (response.data && response.data.event) {
      debug('Received event data:', {
        payload: response.data.event.payload,
        headers: response.data.event.headers,
      });
    }

    expect(response.data.success).to.be.equal(true);

    if (encoding === 'binary' || encoding === 'legacy') {
      expect(response.data).to.have.nested.property('event.payload.eventId',
          eventId, 'The same event id expected in the result');

      // comparing the unclean event type from payload.
      // In headers the event type would be clean one, so comparison may fail.
      expect(response.data).to.have.nested.property(
          'event.payload.eventType', eventType, 'The same event type expected in the result');
    } else if (encoding === 'structured') {
      expect(response.data).to.have.nested.property('event.headers.ce-eventid',
          eventId, 'The same event id expected in the result');

      expect(response.data).to.have.nested.property(
          'event.headers.ce-eventtype', eventType, 'The same event type expected in the result');
    } else {
      throw new Error('Invalid encoding. Possible values are [binary, structured, legacy]');
    }

    return {
      response,
    };
  }, retriesLeft, 2 * 1000)
      .catch((err) => {
        throw convertAxiosError(err, 'Fetching published event responded with error');
      });
}

function createBinaryCloudEventRequestBody(eventId, eventType, eventSource, traceParent = '') {
  debug('setting headers and payload for binary cloud event');
  const reqBody = {
    url: `http://${eppInClusterUrl}/publish`,
    data: {},
  };

  reqBody.data.headers = {
    'ce-source': eventSource,
    'ce-specversion': '1.0',
    'ce-eventtypeversion': 'v1',
    'ce-id': eventId,
    'ce-type': eventType,
    'Content-Type': 'application/json',
    'traceparent': traceParent,
  };

  reqBody.data.payload = {
    eventId: eventId,
    eventType: eventType, // passing unclean event type as payload
  };
  return reqBody;
}

function createStructuredCloudEventRequestBody(eventId, eventType, eventSource, traceparent) {
  debug('setting headers and payload for structured cloud event');
  const reqBody = {
    url: `http://${eppInClusterUrl}/publish`,
    data: {},
  };

  reqBody.data.headers = {
    'Content-Type': 'application/cloudevents+json',
    'traceparent': traceparent,
  };

  reqBody.data.payload = {
    source: eventSource,
    specversion: '1.0',
    eventtypeversion: 'v1',
    datacontenttype: 'application/json',
    id: eventId,
    type: eventType,
    eventId: eventId,
    eventType: eventType, // passing unclean event type as payload
    data: '{"foo":"bar"}',
  };
  return reqBody;
}

function createLegacyEventRequestBody(eventId, eventType, eventSource, isSubV1Alpha1 = true) {
  debug('setting url, headers and payload for legacy event');
  // event types are different between subscription v1alpha1 and v1alpha2.
  // so extracting the appropriate types for legacy format.
  let legacyVersion; let legacySource; let legacyType;
  if (isSubV1Alpha1) {
    // e.g. type: sap.kyma.custom.noapp.order.created.v1
    const typeSegments = eventType.replace('sap.kyma.custom.', '').split('.');
    // extract source (e.g. noapp)
    legacySource = typeSegments[0];
    // extract last version info (e.g. v1)
    legacyVersion = typeSegments[typeSegments.length-1];
    // remove last version (e.g. order.created)
    legacyType = typeSegments.slice(1, typeSegments.length-1).join('.');
  } else {
    // e.g. type: order.created.v1
    const typeSegments = eventType.split('.');

    // extract last version info (e.g. v1)
    legacyVersion = typeSegments[typeSegments.length-1];
    // remove last version (e.g. order.created)
    legacyType = typeSegments.slice(0, typeSegments.length-1).join('.');
    legacySource = eventSource;
  }

  // Now, create the request body
  // Note that EPP publish URL is different for legacy events
  const reqBody = {
    url: `http://${eppInClusterUrl}/${legacySource}/v1/events`,
    data: {},
  };

  reqBody.data.headers = {
    'Content-Type': 'application/json',
  };

  reqBody.data.payload = {
    'event-id': eventId,
    'event-type': legacyType,
    'event-source': legacySource,
    'event-type-version': legacyVersion,
    'event-time': '2020-09-28T14:47:16.491Z',
    'data': {
      eventId: eventId,
      eventType: eventType, // passing unclean event type as payload
    },
  };
  return reqBody;
}

async function getConfigMapWithRetries(name, namespace, retriesLeft = 10) {
  return retryPromise(async () => {
    try {
      return await getConfigMap(name, namespace);
    } catch (err) {
      if (err.statusCode === 404) {
        return undefined;
      }
      throw err;
    }
  }, retriesLeft, 1000);
}

async function createK8sConfigMapWithRetries(data, name, namespace, retriesLeft = 10) {
  return retryPromise(async () => createK8sConfigMap(data, name, namespace), retriesLeft, 1000);
}

async function getJetStreamStreamDataV2(host, streamName) {
  const responseJson = await retryPromise(async () => await axios.get(`https://${host}/jsz?streams=true`), 5, 1000);
  const streams = responseJson.data.account_details[0].stream_detail;
  for (const stream of streams) {
    if (stream.name === streamName) {
      return stream;
    }
  }
  return undefined;
}

async function getJetStreamConsumerDataV2(consumerName, host) {
  const responseJson = await retryPromise(async () => await axios.get(`https://${host}/jsz?consumers=true`), 5, 1000);
  const consumers = responseJson.data.account_details[0].stream_detail[0].consumer_detail;
  for (const consumer of consumers) {
    if (consumer.name === consumerName) {
      return consumer;
    }
  }
  return undefined;
}

async function saveJetStreamDataForRecreateTest(host, configMapName) {
  debug('Fetching stream details from NATS server...');
  const streamData = await getJetStreamStreamDataV2(host, kymaStreamName);
  expect(streamData).to.not.be.undefined;

  const consumerData = {};
  debug(`Fetching consumer (${subscriptionsTypes[0].consumerName}) details from NATS...`);
  const consumerInfo = await getJetStreamConsumerDataV2(subscriptionsTypes[0].consumerName, host);
  expect(consumerInfo).to.not.be.undefined;
  // save by consumer name as key
  consumerData[subscriptionsTypes[0].consumerName] = consumerInfo;

  // Note that the values for stringified.
  const cmData = {
    stream: JSON.stringify(streamData),
    consumers: JSON.stringify(consumerData),
  };

  debug(`Saving fetched stream and consumers details in configMap (name: ${configMapName})...`);
  await createK8sConfigMapWithRetries(cmData, configMapName, testNamespace);
}

async function checkStreamNotReCreated(host, preUpgradeStreamData) {
  debug('Fetching latest stream details from NATS server...');
  const streamData = await getJetStreamStreamDataV2(host, kymaStreamName);
  expect(streamData).to.not.be.undefined;

  const beforeUpgradeCreationTime = preUpgradeStreamData.created;
  const afterUpgradeCreationTime = streamData.created;

  debug(`Stream creation timestamp: 
    Before Upgrade: ${beforeUpgradeCreationTime}, After Upgrade: ${afterUpgradeCreationTime}`);
  expect(getTimeStampsWithZeroMilliSeconds(beforeUpgradeCreationTime)).to.be.equal(
      getTimeStampsWithZeroMilliSeconds(afterUpgradeCreationTime));
}

async function checkConsumerNotReCreated(host, preUpgradeConsumersData) {
  const consumerName = subscriptionsTypes[0].consumerName;
  expect(preUpgradeConsumersData[consumerName]).to.not.be.undefined;

  debug(`Fetching consumer (name: ${consumerName}) latest details from NATS server...`);
  const consumerInfo = await getJetStreamConsumerDataV2(consumerName, host);
  expect(consumerInfo).to.not.be.undefined;

  const beforeUpgradeCreationTime = preUpgradeConsumersData[consumerName].created;
  const afterUpgradeCreationTime = consumerInfo.created;
  debug(`Consumer creation timestamp: 
    Before Upgrade: ${beforeUpgradeCreationTime}, After Upgrade: ${afterUpgradeCreationTime}`);
  expect(getTimeStampsWithZeroMilliSeconds(beforeUpgradeCreationTime)).to.be.equal(
      getTimeStampsWithZeroMilliSeconds(afterUpgradeCreationTime));
}

function getTimeStampsWithZeroMilliSeconds(timestamp) {
  // set milliseconds to zero
  const ts = (new Date(timestamp)).setMilliseconds(0);
  return (new Date(ts)).toISOString();
}

async function createK8sNamespace(name) {
  await k8sApplyWithRetries([namespaceObj(name)]);
}

function debugBanner(message) {
  const line = '[BANNER] ***************************************************************************************';
  debug(line);
  debug(`[BANNER] ${message}`);
  debug(line);
}

module.exports = {
  appName,
  testNamespace,
  kymaVersion,
  isSKR,
  isUpgradeJob,
  skrInstanceId,
  testCompassFlow,
  backendK8sSecretName,
  backendK8sSecretNamespace,
  testDataConfigMapName,
  jsRecreatedTestConfigMapName,
  eventingNatsSvcName,
  eventingNatsApiRuleAName,
  timeoutTime,
  slowTime,
  director,
  gardener,
  shootName,
  suffix,
  eppInClusterUrl,
  eventingSinkName,
  eventingUpgradeSinkName,
  v1alpha1SubscriptionsTypes,
  subscriptionsTypes,
  subscriptionsExactTypeMatching,
  kymaStreamName,
  getRegisteredCompassScenarios,
  ensureEventReceivedWithRetry,
  cleanupTestingResources,
  getClusterHost,
  checkFunctionReachable,
  checkFunctionUnreachable,
  checkEventDelivery,
  deployEventingSinkFunction,
  undeployEventingFunction,
  waitForEventingSinkFunction,
  deployV1Alpha1Subscriptions,
  deployV1Alpha2Subscriptions,
  waitForV1Alpha1Subscriptions,
  waitForV1Alpha2Subscriptions,
  checkEventTracing,
  saveJetStreamDataForRecreateTest,
  getConfigMapWithRetries,
  checkStreamNotReCreated,
  checkConsumerNotReCreated,
  createK8sNamespace,
  publishEventWithRetry,
  debugBanner,
  isJSRecreatedTestEnabled,
  isJSAtLeastOnceDeliveryTestEnabled,
};
