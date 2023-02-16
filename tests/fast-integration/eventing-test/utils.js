const {
  cleanMockTestFixture,
  cleanCompassResourcesSKR,
  generateTraceParentHeader,
  checkTrace,
} = require('../test/fixtures/commerce-mock');

const {
  debug,
  getEnvOrThrow,
  deleteEventingBackendK8sSecret,
  deleteK8sConfigMap,
  getShootNameFromK8sServerUrl,
  listPods,
  retryPromise,
  waitForVirtualService,
  k8sApply,
  waitForFunction,
  eventingSubscription,
  waitForSubscription,
  eventingSubscriptionV1Alpha2,
  getSecretData,
  convertAxiosError,
  sleep,
} = require('../utils');

const {DirectorClient, DirectorConfig, getAlreadyAssignedScenarios} = require('../compass');
const {GardenerClient, GardenerConfig} = require('../gardener');
const {eventMeshSecretFilePath} = require('./common/common');
const axios = require('axios');
const {v4: uuidv4} = require('uuid');
const fs = require('fs');
const path = require('path');
const k8s = require('@kubernetes/client-node');
const {OAuthToken, OAuthCredentials} = require('../lib/oauth');
const {expect} = require('chai');

// Variables
const kymaVersion = process.env.KYMA_VERSION || '';
const isSKR = process.env.KYMA_TYPE === 'SKR';
const skrInstanceId = process.env.INSTANCE_ID || '';
const testCompassFlow = process.env.TEST_COMPASS_FLOW || false;
const testSubscriptionV1Alpha2 = process.env.ENABLE_SUBSCRIPTION_V1_ALPHA2 === 'true';
const subCRDVersion = testSubscriptionV1Alpha2? 'v1alpha2': 'v1alpha1';
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
const eppInClusterUrl = 'eventing-event-publisher-proxy.kyma-system';
const subscriptionNames = {
  orderCreated: 'order-created',
  orderReceived: 'order-received',
};
const eventingSinkName = 'eventing-sink';

// ****** Event types to test ***********//
const v1alpha1SubscriptionsTypes = [
  'sap.kyma.custom.noapp.order.tested.v1',
  'sap.kyma.custom.connected-app.order.tested.v1',
  'sap.kyma.custom.test-app.order-$.second.R-e-c-e-i-v-e-d.v1',
  'sap.kyma.custom.connected-app2.or-der.crea-ted.one.two.three.v4',
];

const subscriptionsTypes = [
  {
    type: 'order.modified.v1',
    source: 'myapp',
  },
  {
    type: 'or-der.crea-ted.one.two.three.four.v4',
    source: 'test-app',
  },
  {
    type: 'Order-$.third.R-e-c-e-i-v-e-d.v1',
    source: 'test-app',
  },
];

// ****** ************* ***********//

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

function isStreamCreationTimeMissing(streamData) {
  return streamData.streamCreationTime === undefined;
}

async function getClusterHost(apiRuleName, namespace) {
  const vs = await waitForVirtualService(namespace, apiRuleName);
  const mockHost = vs.spec.hosts[0];
  return mockHost.split('.').slice(1).join('.');
}

function createNewEventId() {
  return uuidv4();
}

async function deployEventingSinkFunction() {
  const functionYaml = fs.readFileSync(
      path.join(__dirname, './assets/eventing-function.yaml'),
      {
        encoding: 'utf8',
      },
  );

  const k8sObjs = k8s.loadAllYaml(functionYaml);
  await k8sApply(k8sObjs, testNamespace, true);
}

async function waitForEventingSinkFunction() {
  await waitForFunction(eventingSinkName, testNamespace);
}

async function deployV1Alpha1Subscriptions() {
  const sink = `http://${eventingSinkName}.${testNamespace}.svc.cluster.local`;
  debug(`Using sink: ${sink}`);

  // creating v1alpha1 subscriptions
  for (let i=0; i < v1alpha1SubscriptionsTypes.length; i++) {
    const subName = `fi-test-sub-${i}`;
    const eventType = v1alpha1SubscriptionsTypes[i];

    debug(`Creating subscription: ${subName} with type: ${eventType}`);
    await k8sApply([eventingSubscription(eventType, sink, subName, testNamespace)]);
    debug(`Waiting for subscription: ${subName} with type: ${eventType}`);
    await waitForSubscription(subName, testNamespace);
  }
}

async function deployV1Alpha2Subscriptions() {
  const sink = `http://${eventingSinkName}.${testNamespace}.svc.cluster.local`;
  debug(`Using sink: ${sink}`);

  // creating v1alpha2 subscriptions
  for (let i=0; i < subscriptionsTypes.length; i++) {
    const subName = `fi-test-sub-v2-${i}`;
    const eventType = subscriptionsTypes[i].type;
    const eventSource = subscriptionsTypes[i].source;

    debug(`Creating subscription: ${subName} with type: ${eventType}, source: ${eventSource}`);
    // eventingSubscriptionV1Alpha2(eventType, source, sink, name, namespace, typeMatching='standard')
    await k8sApply([eventingSubscriptionV1Alpha2(eventType, eventSource, sink, subName, testNamespace)]);
    debug(`Waiting for subscription: ${subName} with type: ${eventType}, source: ${eventSource}`);
    await waitForSubscription(subName, testNamespace);
  }
}

async function waitForV1Alpha1Subscriptions() {
  // waiting for v1alpha1 subscriptions
  for (let i=0; i < v1alpha1SubscriptionsTypes.length; i++) {
    const subName = `fi-test-sub-${i}`;
    debug(`Waiting for subscription: ${subName} with type: ${v1alpha1SubscriptionsTypes[i]}`);
    await waitForSubscription(subName, testNamespace);
  }
}

async function waitForV1Alpha2Subscriptions() {
  // waiting for v1alpha2 subscriptions
  for (let i=0; i < subscriptionsTypes.length; i++) {
    const subName = `fi-test-sub-v2-${i}`;
    debug(`Waiting for subscription: ${subName} with type: ${subscriptionsTypes[i].type}`);
    await waitForSubscription(subName, testNamespace);
  }
}

async function checkFunctionReachable(name, namespace, host) {
  // get OAuth client id and client secret from Kubernetes Secret
  const oAuthSecretData = await getSecretData(`${name}-oauth`, namespace);

  // get access token from OAuth server
  const oAuthTokenGetter = new OAuthToken(
      `https://oauth2.${host}/oauth2/token`,
      new OAuthCredentials(oAuthSecretData['client_id'], oAuthSecretData['client_secret']),
  );
  const accessToken = await oAuthTokenGetter.getToken(['read', 'write']);

  // expect no error when authorized
  const res = await retryPromise(
      () => axios.post(`https://${name}.${host}/function`, {orderCode: '789'}, {
        timeout: 5000,
        headers: {Authorization: `bearer ${accessToken}`},
      }),
      45,
      2000,
  ).catch((err) => {
    throw convertAxiosError(err, `Function ${name} responded with error`);
  });

  // the request should be authorized and successful
  expect(res.status).to.be.equal(200);

  // expect error when unauthorized
  let errorOccurred = false;
  try {
    await axios.post(`https://${name}.${host}/function`, {orderCode: '789'}, {timeout: 5000});
  } catch (err) {
    errorOccurred = true;
    expect(err.response.status).to.be.equal(401);
  }
  expect(errorOccurred).to.be.equal(true);
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
async function checkEventDelivery(proxyHost, encoding, eventType, eventSource, isSubV1Alpha1 = false) {
  const eventId = createNewEventId();

  debug(`Publishing event with id:${eventId}, type: ${eventType}, source: ${eventSource}...`);
  const result = await publishEventWithRetry(proxyHost, encoding, eventId, eventType, eventSource, isSubV1Alpha1);

  debug(`Verifying if event with id:${eventId}, type: ${eventType}, source: ${eventSource} was received by sink...`);
  const result2 = await ensureEventReceivedWithRetry(proxyHost, encoding, eventId, eventType, eventSource);
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
      reqBody = createBinaryCloudEventRequestBody(proxyHost, eventId, eventType, eventSource, traceParentId);
    } else if (encoding === 'structured') { // structured CE
      reqBody = createStructuredCloudEventRequestBody(proxyHost, eventId, eventType, eventSource, traceParentId);
    } else if (encoding === 'legacy') {
      reqBody = createLegacyEventRequestBody(proxyHost, eventId, eventType, eventSource, isSubV1Alpha1);
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
async function ensureEventReceivedWithRetry(proxyHost, encoding, eventId, eventType, eventSource, retriesLeft = 10) {
  return await retryPromise(async () => {
    debug(`Waiting to receive CE event "${eventId}"`);

    const response = await axios.get(`https://${eventingSinkName}.${proxyHost}`,
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
        ceHeaders: response.data.event.ceHeaders,
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

function createBinaryCloudEventRequestBody(proxyHost, eventId, eventType, eventSource, traceparent) {
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
    'traceparent': traceparent,
  };

  reqBody.data.payload = {
    eventId: eventId,
    eventType: eventType, // passing unclean event type as payload
  };
  return reqBody;
}

function createStructuredCloudEventRequestBody(proxyHost, eventId, eventType, eventSource, traceparent) {
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
  };
  return reqBody;
}

function createLegacyEventRequestBody(proxyHost, eventId, eventType, eventSource, isSubV1Alpha1 = true) {
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

module.exports = {
  appName,
  scenarioName,
  testNamespace,
  mockNamespace,
  kymaVersion,
  isSKR,
  skrInstanceId,
  testCompassFlow,
  testSubscriptionV1Alpha2,
  subCRDVersion,
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
  isStreamCreationTimeMissing,
  eppInClusterUrl,
  subscriptionNames,
  eventingSinkName,
  v1alpha1SubscriptionsTypes,
  subscriptionsTypes,
  getClusterHost,
  checkFunctionReachable,
  checkEventDelivery,
  deployEventingSinkFunction,
  waitForEventingSinkFunction,
  deployV1Alpha1Subscriptions,
  deployV1Alpha2Subscriptions,
  waitForV1Alpha1Subscriptions,
  waitForV1Alpha2Subscriptions,
  checkEventTracing,
};
