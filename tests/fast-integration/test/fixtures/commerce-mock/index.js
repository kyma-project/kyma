const k8s = require('@kubernetes/client-node');
const fs = require('fs');
const path = require('path');
const {expect} = require('chai');
const https = require('https');
const axios = require('axios').default;
const httpsAgent = new https.Agent({
  rejectUnauthorized: false, // curl -k
});
axios.defaults.httpsAgent = httpsAgent;

const {
  retryPromise,
  convertAxiosError,
  sleep,
  k8sApply,
  waitForApplicationCr,
  waitForVirtualService,
  waitForDeployment,
  waitForFunction,
  waitForSubscription,
  deleteAllK8sResources,
  genRandom,
  k8sDynamicApi,
  deleteNamespaces,
  error,
  debug,
  isDebugEnabled,
  toBase64,
  eventingSubscription,
  k8sDelete,
  getSecretData,
  namespaceObj,
  getTraceDAG,
  printStatusOfInClusterEventingInfrastructure,
  labelNamespaceWithIstioInject,
} = require('../../../utils');

const {
  registerOrReturnApplication,
  deregisterApplication,
  removeApplicationFromScenario,
  removeScenarioFromCompass,
  getApplicationByName,
  unassignRuntimeFromScenario,
} = require('../../../compass');

const {getJaegerTrace} = require('../../../tracing/client');

const {
  OAuthToken,
  OAuthCredentials,
} = require('../../../lib/oauth');

const {bebBackend, getEventMeshNamespace} = require('../../../eventing-test/common/common');

const commerceMockYaml = fs.readFileSync(
    path.join(__dirname, './commerce-mock.yaml'),
    {
      encoding: 'utf8',
    },
);

const applicationMockYaml = fs.readFileSync(
    path.join(__dirname, './application.yaml'),
    {
      encoding: 'utf8',
    },
);

const lastorderFunctionYaml = fs.readFileSync(
    path.join(__dirname, './lastorder-function.yaml'),
    {
      encoding: 'utf8',
    },
);

const applicationObjs = k8s.loadAllYaml(applicationMockYaml);
const lastorderObjs = k8s.loadAllYaml(lastorderFunctionYaml);
let eventMeshSourceNamespace = '/default/sap.kyma/tunas-prow';

function setEventMeshSourceNamespace(namespace) {
  eventMeshSourceNamespace = `/${namespace.trimStart('/')}`;
}

function prepareFunction(type = 'standard', appName = 'commerce') {
  const functionYaml = lastorderFunctionYaml.toString().replace(/%%BEB_NAMESPACE%%/g, eventMeshSourceNamespace);
  const gatewayUrl = 'http://central-application-gateway.kyma-system';
  switch (type) {
    case 'central-app-gateway':
      const orders = `${gatewayUrl}:8080/commerce/sap-commerce-cloud-commerce-webservices/site/orders/`;
      return k8s.loadAllYaml(functionYaml.toString()
          .replace('%%URL%%', `"${orders}" + code`));
    case 'central-app-gateway-compass':
      const ordersWithCompass = `${gatewayUrl}:8082/%%APP_NAME%%/sap-commerce-cloud/commerce-webservices/site/orders/`;
      return k8s.loadAllYaml(functionYaml.toString()
          .replace('%%URL%%', `"${ordersWithCompass}" + code`)
          .replace('%%APP_NAME%%', appName));
    default:
      return k8s.loadAllYaml(functionYaml.toString()
          .replace('%%URL%%', 'findEnv("GATEWAY_URL") + "/site/orders/" + code'));
  }
}

// Allows creating Commerce Mock objects in a specific namespace
function prepareCommerceObjs(mockNamespace) {
  return k8s.loadAllYaml(commerceMockYaml.toString().replace(/%%MOCK_NAMESPACE%%/g, mockNamespace));
}

async function checkFunctionResponse(functionNamespace, mockNamespace = 'mocks') {
  const vs = await waitForVirtualService(mockNamespace, 'commerce-mock');
  const mockHost = vs.spec.hosts[0];
  const host = mockHost.split('.').slice(1).join('.');

  // get OAuth client id and client secret from Kubernetes Secret
  const oAuthSecretData = await getSecretData('lastorder-oauth', functionNamespace);

  // get access token from OAuth server
  const oAuthTokenGetter = new OAuthToken(
      `https://oauth2.${host}/oauth2/token`,
      new OAuthCredentials(oAuthSecretData['client_id'], oAuthSecretData['client_secret']),
  );
  const accessToken = await oAuthTokenGetter.getToken(['read', 'write']);

  // expect no error when authorized
  const res = await retryPromise(
      () => axios.post(`https://lastorder.${host}/function`, {orderCode: '789'}, {
        timeout: 5000,
        headers: {Authorization: `bearer ${accessToken}`},
      }),
      45,
      2000,
  ).catch((err) => {
    throw convertAxiosError(err, 'Function lastorder responded with error');
  });

  // the request should be authorized and successful
  expect(res.status).to.be.equal(200);

  // expect error when unauthorized
  let errorOccurred = false;
  try {
    await axios.post(`https://lastorder.${host}/function`, {orderCode: '789'}, {timeout: 5000});
  } catch (err) {
    errorOccurred = true;
    expect(err.response.status).to.be.equal(401);
  }
  expect(errorOccurred).to.be.equal(true);
}

async function sendEventAndCheckResponse(eventType, body, params, mockNamespace = 'mocks') {
  const vs = await waitForVirtualService(mockNamespace, 'commerce-mock');
  const mockHost = vs.spec.hosts[0];
  const host = mockHost.split('.').slice(1).join('.');

  return await retryPromise(
      async () => {
        await axios
            .post(`https://${mockHost}/events`, body, params)
            .catch((e) => {
              error('Cannot send %s, the response from event gateway: %s', eventType, e.response.data);
              console.log(e);
              throw convertAxiosError(e, 'Cannot send %s, the response from event gateway', eventType);
            });

        await sleep(500);

        return axios
            .get(`https://lastorder.${host}`, {timeout: 5000})
            .then((res) => {
              expect(res.data).to.have.nested.property('event.data.orderCode', '567');
              // See: https://github.com/kyma-project/kyma/issues/10720
              expect(res.data).to.have.nested.property('event.ce-type').that.contains('order.created');
              expect(res.data).to.have.nested.property('event.ce-source');
              expect(res.data).to.have.nested.property('event.ce-eventtypeversion', 'v1');
              expect(res.data).to.have.nested.property('event.ce-specversion', '1.0');
              expect(res.data).to.have.nested.property('event.ce-id');
              expect(res.data).to.have.nested.property('event.ce-time');
              return res;
            })
            .catch((e) => {
              throw convertAxiosError(e, 'Error during request to function lastorder');
            });
      },
      10,
      30 * 1000,
  );
}

async function sendLegacyEventAndCheckResponse(mockNamespace = 'mocks') {
  const body = {
    'event-type': 'order.created',
    'event-type-version': 'v1',
    'event-time': '2020-09-28T14:47:16.491Z',
    'data': {'orderCode': '567'},
    // this parameter sets the x-b3-sampled header on the commerce-mock side,
    // which configures istio-proxies to collect the traces no matter what sampling rate is configured
    'event-tracing': true,
  };
  const params = {
    headers: {
      'content-type': 'application/json',
    },
  };

  return await sendEventAndCheckResponse('legacy event', body, params, mockNamespace);
}

async function sendCloudEventStructuredModeAndCheckResponse(backendType ='nats', mockNamespace = 'mocks') {
  let source = 'commerce';
  if (backendType === bebBackend) {
    source = getEventMeshNamespace();
  }
  const body = {
    'specversion': '1.0',
    'source': source,
    'type': 'sap.kyma.custom.noapp.order.created.v1',
    'eventtypeversion': 'v1',
    'id': 'A234-1234-1234',
    'data': {'orderCode': '567'},
    'datacontenttype': 'application/json',
    'eventtracing': true,
  };
  const params = {
    headers: {
      'content-type': 'application/cloudevents+json',
    },
  };

  return await sendEventAndCheckResponse('cloud event', body, params, mockNamespace);
}

async function sendCloudEventBinaryModeAndCheckResponse(backendType = 'nats', mockNamespace = 'mocks') {
  let source = 'commerce';
  if (backendType === bebBackend) {
    source = getEventMeshNamespace();
  }
  const body = {
    'data': {'orderCode': '567'},
    'eventtracing': true,
  };
  const params = {
    headers: {
      'content-type': 'application/json',
      'ce-specversion': '1.0',
      'ce-type': 'sap.kyma.custom.noapp.order.created.v1',
      'ce-source': source,
      'ce-id': 'A234-1234-1234',
    },
  };

  return await sendEventAndCheckResponse('cloud event binary', body, params, mockNamespace);
}

async function checkEventTracing(targetNamespace = 'test', res) {
  expect(res.data).to.have.nested.property('event.headers.x-b3-traceid');
  expect(res.data).to.have.nested.property('podName');

  // Extract traceId from response
  const traceId = res.data.event.headers['x-b3-traceid'];

  // Define expected trace data
  const correctTraceProcessSequence = [
    'istio-ingressgateway.istio-system',
    'central-application-connectivity-validator.kyma-system',
    'central-application-connectivity-validator.kyma-system',
    'eventing-publisher-proxy.kyma-system',
    'eventing-controller.kyma-system',
    `lastorder-${res.data.podName.split('-')[1]}.${targetNamespace}`,
  ];
  // wait some time for jaeger to complete tracing data
  await sleep(10 * 1000);
  await checkTrace(traceId, correctTraceProcessSequence);
}

async function sendLegacyEventAndCheckTracing(targetNamespace = 'test', mockNamespace = 'mocks') {
  // Send an event and get it back from the lastorder function
  const res = await sendLegacyEventAndCheckResponse(mockNamespace);

  // Check the correct event tracing
  await checkEventTracing(targetNamespace, res);
}

async function sendCloudEventStructuredModeAndCheckTracing(targetNamespace = 'test', mockNamespace = 'mocks') {
  // Send an event and get it back from the lastorder function
  const res = await sendCloudEventStructuredModeAndCheckResponse(mockNamespace);

  // Check the correct event tracing
  await checkEventTracing(targetNamespace, res);
}

async function sendCloudEventBinaryModeAndCheckTracing(targetNamespace = 'test', mockNamespace = 'mocks') {
  // Send an event and get it back from the lastorder function
  const res = await sendCloudEventBinaryModeAndCheckResponse(mockNamespace);

  // Check the correct event tracing
  await checkEventTracing(targetNamespace, res);
}

async function checkInClusterEventTracing(targetNamespace) {
  const res = await checkInClusterEventDeliveryHelper(targetNamespace, 'structured');
  expect(res.data).to.have.nested.property('event.headers.x-b3-traceid');
  expect(res.data).to.have.nested.property('podName');

  // Extract traceId from response
  const traceId = res.data.event.headers['x-b3-traceid'];

  // Define expected trace data
  const correctTraceProcessSequence = [
    // We are sending the in-cluster event from inside the lastorder pod
    'istio-ingressgateway.istio-system',
    `lastorder-${res.data.podName.split('-')[1]}.${targetNamespace}`,
    'eventing-publisher-proxy.kyma-system',
    'eventing-controller.kyma-system',
    `lastorder-${res.data.podName.split('-')[1]}.${targetNamespace}`,
  ];

  // wait sometime for jaeger to complete tracing data
  await sleep(10 * 1000);
  await checkTrace(traceId, correctTraceProcessSequence);
}

async function checkTrace(traceId, expectedTraceProcessSequence) {
  const traceRes = await getJaegerTrace(traceId);

  // the trace response should have data for single trace
  expect(traceRes.data).to.have.length(1);

  // extract trace data from response
  const traceData = traceRes.data[0];
  expect(traceData['spans'].length).to.be.gte(expectedTraceProcessSequence.length);

  // generate DAG for trace spans
  const traceDAG = await getTraceDAG(traceData);
  expect(traceDAG).to.have.length(1);

  // searching through the trace-graph for the expected span sequence staring at the root element
  const wasFound = findSpanSequence(expectedTraceProcessSequence, 0, traceDAG[0], traceData);
  if (!wasFound) {
    debug(`Not all expected spans found in the expected order:`);
    for (let i = 0; i < expectedTraceProcessSequence.length; i++) {
      debug(`${expectedTraceProcessSequence[i]}`);
    }
  }
  expect(wasFound).to.be.true;
}

// findSpanSequence recursively searches through the trace-graph to find all expected spans in the right, consecutive
// order while ignoring the spans that are not expected.
function findSpanSequence(expectedSpans, position, currentSpan, traceData) {
  // validate if the actual span is the expected span
  const actualSpan = traceData.processes[currentSpan.processID].serviceName;
  const expectedSpan = expectedSpans[position];
  let newPosition = position;
  const debugMsg = `${buildLevel(position)} ${actualSpan}`;
  // if this span contains the currently expected span, the position will be increased
  if (actualSpan === expectedSpan) {
    newPosition++;
    debug(debugMsg);
  } else {
    debug(`${debugMsg} expected ${expectedSpan}`);
  }

  // check if all traces have been found yet
  if (newPosition === expectedSpans.length) {
    return true;
  }

  // recursive search through all the child spans
  for (let i = 0; i < currentSpan.childSpans.length; i++) {
    if (findSpanSequence(expectedSpans, newPosition, currentSpan.childSpans[i], traceData)) {
      return true;
    }
  }

  // if nothing was found on this branch of the graph, close it
  return false;
}

// buildLevel helps to display trace hierarchy by adding a whitespace for each level of hierarchy in front of the trace
// to get output like
// -> myTrace
//  └> myChildTrace
//   └> ChildOfMyChildTrace
// ...
function buildLevel(n) {
  if (n === 0) {
    return '  ->';
  }

  let level = '';
  for (let i = 0; i < n+1; i++) {
    level += ' ';
  }
  return `${level} └>`;
}

async function addService() {
  const vs = await waitForVirtualService('mocks', 'commerce-mock');
  const mockHost = vs.spec.hosts[0];
  const url = `https://${mockHost}/remote/apis`;
  const body = {
    'name': 'my-service-http-bin',
    'provider': 'myCompany',
    'description': 'This is some service',
    'api': {
      'targetUrl': 'https://httpbin.org',
      'spec': {
        'swagger': '2.0',
      },
    },
  };
  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
    timeout: 5000,
  };

  let serviceId;
  try {
    serviceId = await axios.post(url, body, params);
  } catch (err) {
    throw convertAxiosError(err, 'Error during adding a Service');
  }
  return serviceId.data.id;
}

async function updateService(serviceId) {
  const vs = await waitForVirtualService('mocks', 'commerce-mock');
  const mockHost = vs.spec.hosts[0];
  const url = `https://${mockHost}/remote/apis/${serviceId}`;
  const body = {
    'name': 'my-service-http-bin',
    'provider': 'myCompany',
    'description': 'This is some service - an updated description',
    'api': {
      'targetUrl': 'https://httpbin.org',
      'spec': {
        'swagger': '2.0',
      },
    },
  };
  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
    timeout: 5000,
  };

  try {
    await axios.put(url, body, params);
  } catch (err) {
    throw convertAxiosError(err, 'Error during updating a Service');
  }
}

async function deleteService(serviceId) {
  const vs = await waitForVirtualService('mocks', 'commerce-mock');
  const mockHost = vs.spec.hosts[0];
  const url = `https://${mockHost}/remote/apis/${serviceId}`;
  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
    timeout: 5000,
  };

  try {
    await axios.delete(url, params);
  } catch (err) {
    throw convertAxiosError(err, 'Error during deleting a Service');
  }
}

async function registerAllApis(mockHost) {
  debug('Listing Commerce Mock local APIs');
  const localApis = await axios.get(`https://${mockHost}/local/apis`, {timeout: 5000}).catch((err) => {
    throw convertAxiosError(err, 'API registration error - commerce mock local API not available');
  });
  debug('Commerce Mock local APIs received');
  const filteredApis = localApis.data
      .filter((api) => (api.name.includes('Commerce Webservices') || api.name.includes('Events')));
  for (const api of filteredApis) {
    debug('Registering', api.name);
    await axios
        .post(
            `https://${mockHost}/local/apis/${api.id}/register`,
            {},
            {
              headers: {
                'content-type': 'application/json',
                'origin': `https://${mockHost}`,
              },
              timeout: 30000,
            },
        ).catch((err) => {
          throw convertAxiosError(err, 'Error during Commerce Mock API registration');
        });
  }
  debug('Verifying if APIs are properly registered');

  const remoteApis = await axios
      .get(`https://${mockHost}/remote/apis`)
      .catch((err) => {
        throw convertAxiosError(err, 'Commerce Mock registered apis not available');
      });
  expect(remoteApis.data).to.have.lengthOf.at.least(2);
  debug('Commerce APIs registered');
  return remoteApis;
}

async function connectMockCompass(client, appName, scenarioName, mockHost, targetNamespace) {
  const appID = await registerOrReturnApplication(client, appName, scenarioName);
  debug(`Application ID in Compass ${appID}`);

  const pairingData = await client.requestOneTimeTokenForApplication(appID);
  const pairingToken = toBase64(JSON.stringify(pairingData));
  const pairingBody = {
    token: pairingToken,
    baseUrl: mockHost,
    insecure: true,
  };

  debug(`Connecting ${mockHost}`);
  await connectCommerceMock(mockHost, pairingBody);

  debug('Commerce mock connected to Compass');
}

async function connectCommerceMock(mockHost, tokenData) {
  const url = `https://${mockHost}/connection`;
  const body = tokenData;
  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
    timeout: 30 * 1000,
  };

  try {
    await axios.post(url, body, params);
  } catch (err) {
    throw convertAxiosError(err, 'Error during establishing connection from Commerce Mock to Kyma connector service');
  }
}

async function ensureCommerceMockWithCompassTestFixture(
    client,
    appName,
    scenarioName,
    mockNamespace,
    targetNamespace,
    compassScenarioAlreadyExist = false) {
  const lastOrderFunction = prepareFunction('central-app-gateway-compass', `mp-${appName}`);

  const mockHost = await provisionCommerceMockResources(
      `mp-${appName}`,
      mockNamespace,
      targetNamespace,
      lastOrderFunction);
  await retryPromise(() => connectMockCompass(client, appName, scenarioName, mockHost, targetNamespace), 10, 30000);
  // do not register the apis again for an already existing compass scenario
  if (!compassScenarioAlreadyExist) {
    await retryPromise(() => registerAllApis(mockHost), 10, 30000);
  }

  await waitForDeployment('central-application-gateway', 'kyma-system');
  await waitForDeployment('central-application-connectivity-validator', 'kyma-system');

  await waitForFunction('lastorder', targetNamespace);

  await waitForApplicationCr(`mp-${appName}`);

  await k8sApply([eventingSubscription(
      `sap.kyma.custom.inapp.order.received.v1`,
      `http://lastorder.${targetNamespace}.svc.cluster.local`,
      'order-received',
      targetNamespace)]);
  await waitForSubscription('order-received', targetNamespace);
  await waitForSubscription('order-created', targetNamespace);
  return mockHost;
}

async function cleanCompassResourcesSKR(client, appName, scenarioName, runtimeID) {
  const application = await getApplicationByName(client, appName, scenarioName);
  if (application) {
    // detach Commerce-mock application from scenario
    // so that we can de-register the app from compass
    console.log(`Removing application from scenario...`);
    await removeApplicationFromScenario(client, application.id, scenarioName);

    // Disconnect Commerce-mock app from compass
    console.log(`De-registering application: ${application.id}...`);
    await deregisterApplication(client, application.id);
  }

  try {
    // detach the target SKR from scenario
    // so that we can remove scenario from compass
    console.log(`Un-assigning runtime from scenario: ${scenarioName}...`);
    await unassignRuntimeFromScenario(client, runtimeID, scenarioName);

    console.log(`Removing scenario from compass: ${scenarioName}...`);
    await removeScenarioFromCompass(client, scenarioName);
  } catch (err) {
    console.log(`Error: Failed to remove scenario from compass`);
    console.log(err);
  }
}

async function ensureCommerceMockLocalTestFixture(mockNamespace, targetNamespace) {
  await k8sApply(applicationObjs);
  const mockHost = await provisionCommerceMockResources(
      'commerce',
      mockNamespace,
      targetNamespace,
      prepareFunction('central-app-gateway'));

  await waitForDeployment('central-application-gateway', 'kyma-system');

  await waitForFunction('lastorder', targetNamespace);

  await k8sApply([eventingSubscription(
      `sap.kyma.custom.inapp.order.received.v1`,
      `http://lastorder.${targetNamespace}.svc.cluster.local`,
      'order-received',
      targetNamespace)]);
  await waitForSubscription('order-received', targetNamespace);
  await waitForSubscription('order-created', targetNamespace);

  return mockHost;
}

async function provisionCommerceMockResources(appName, mockNamespace, targetNamespace, functionObjs = lastorderObjs) {
  await k8sApply([namespaceObj(mockNamespace), namespaceObj(targetNamespace)]);
  await k8sApply(prepareCommerceObjs(mockNamespace));
  await k8sApply(functionObjs, targetNamespace, true);
  await labelNamespaceWithIstioInject(targetNamespace, 'enabled');
  await waitForFunction('lastorder', targetNamespace);
  await k8sApply([
    eventingSubscription(
        `sap.kyma.custom.${appName}.order.created.v1`,
        `http://lastorder.${targetNamespace}.svc.cluster.local`,
        'order-created',
        targetNamespace),
  ]);
  await waitForDeployment('commerce-mock', mockNamespace, 120 * 1000);
  const vs = await waitForVirtualService(mockNamespace, 'commerce-mock');
  const mockHost = vs.spec.hosts[0];
  await retryPromise(
      () =>
        axios.get(`https://${mockHost}/local/apis`).catch((err) => {
          throw convertAxiosError(err, 'Commerce mock local API not available - timeout');
        }),
      40,
      3000,
  );

  return mockHost;
}

function getResourcePaths(namespace) {
  return [
    `/apis/serverless.kyma-project.io/v1alpha1/namespaces/${namespace}/functions`,
    `/apis/addons.kyma-project.io/v1alpha1/namespaces/${namespace}/addonsconfigurations`,
    `/apis/gateway.kyma-project.io/v1alpha1/namespaces/${namespace}/apirules`,
    `/apis/apps/v1/namespaces/${namespace}/deployments`,
    `/api/v1/namespaces/${namespace}/services`,
  ];
}

async function cleanMockTestFixture(mockNamespace, targetNamespace, wait = true) {
  for (const path of getResourcePaths(mockNamespace).concat(
      getResourcePaths(targetNamespace),
  )) {
    await deleteAllK8sResources(path);
  }

  try {
    debug('Deleting applicationconnector.kyma-project.io/v1alpha1');
    await k8sDynamicApi.delete({
      apiVersion: 'applicationconnector.kyma-project.io/v1alpha1',
      kind: 'Application',
      metadata: {
        name: 'commerce',
      },
    });
  } catch (err) {
    // Ignore delete error
  }

  debug('Deleting test namespaces');
  return deleteNamespaces([mockNamespace, targetNamespace], wait);
}

async function deleteMockTestFixture(mockNamespace) {
  await k8sDelete(lastorderObjs);
  await k8sDelete(prepareCommerceObjs(mockNamespace));
  await k8sDelete(applicationObjs);
}

async function waitForSubscriptions(subscriptions) {
  for (let i = 0; i < subscriptions.length; i++) {
    const subscription = subscriptions[i];
    await waitForSubscription(subscription.metadata.name, subscription.metadata.namespace);
  }
}

async function waitForSubscriptionsTillReady(targetNamespace) {
  await waitForSubscription('order-received', targetNamespace);
  await waitForSubscription('order-created', targetNamespace);
}

async function checkInClusterEventDelivery(targetNamespace) {
  await checkInClusterEventDeliveryHelper(targetNamespace, 'structured');
  await checkInClusterEventDeliveryHelper(targetNamespace, 'binary');
  await checkInClusterLegacyEvent(targetNamespace);
}

// send event using function query parameter send=true
async function sendInClusterEventWithRetry(mockHost, eventId, encoding, retriesLeft = 10) {
  await retryPromise(async () => {
    const response = await axios.post(`https://${mockHost}`, {id: eventId}, {
      params: {
        send: true,
        encoding: encoding,
      },
      headers: {
        'X-B3-Sampled': 1,
      },
    });

    debug('Send response:', {
      status: response.status,
      statusText: response.statusText,
      data: response.data,
    });

    if (response.data.eventPublishError) {
      throw convertAxiosError(response.data.statusText);
    }
    expect(response.status).to.be.equal(200);
  }, retriesLeft, 1000);

  debug(`Event "${eventId}" is sent`);
}

// send legacy event using function query parameter send=true
async function sendInClusterLegacyEventWithRetry(mockHost, eventData, retriesLeft = 10) {
  await retryPromise(async () => {
    const response = await axios.post(`https://${mockHost}`, eventData, {
      params: {
        send: true,
        isLegacyEvent: true,
      },
      headers: {
        'X-B3-Sampled': 1,
      },
    });

    debug('Send response:', {
      status: response.status,
      statusText: response.statusText,
      data: response.data,
    });

    if (response.data.eventPublishError) {
      throw convertAxiosError(response.data.statusText);
    }
    expect(response.status).to.be.equal(200);
  }, retriesLeft, 1000);

  debug(`Legacy event is sent: `, eventData);
}

// verify if event was received using function query parameter inappevent=eventId
async function ensureInClusterEventReceivedWithRetry(mockHost, eventId, retriesLeft = 10) {
  return await retryPromise(async () => {
    debug(`Waiting to receive event "${eventId}"`);

    const response = await axios.get(`https://${mockHost}`, {params: {inappevent: eventId}});

    debug('Received response:', {
      status: response.status,
      statusText: response.statusText,
      data: response.data,
    });

    expect(response.data).to.have.nested.property('event.id', eventId, 'The same event id expected in the result');
    expect(response.data).to.have.nested.property('event.shipped', true, 'Order should have property shipped');
    return response;
  }, retriesLeft, 2 * 1000)
      .catch((err) => {
        throw convertAxiosError(err, 'Fetching published event responded with error');
      });
}

// verify if legacy event was received using function query parameter inappevent=eventId
async function ensureInClusterLegacyEventReceivedWithRetry(mockHost, eventId, retriesLeft = 10) {
  return await retryPromise(async () => {
    debug(`Waiting to receive legacy event "${eventId}"`);

    const response = await axios.get(`https://${mockHost}`, {params: {inappevent: eventId}});

    debug('Received response:', {
      status: response.status,
      statusText: response.statusText,
      data: response.data,
    });

    expect(response.data).to.have.nested.property('event.id', eventId, 'The same event id expected in the result');
    expect(response.data).to.have.nested.property('event.shipped', true, 'Order should have property shipped');
    expect(response.data).to.have.nested.property('event.ce-type').that.contains('order.received');
    expect(response.data).to.have.nested.property('event.ce-source');
    expect(response.data).to.have.nested.property('event.ce-eventtypeversion', 'v1');
    expect(response.data).to.have.nested.property('event.ce-specversion', '1.0');
    expect(response.data).to.have.nested.property('event.ce-id');
    expect(response.data).to.have.nested.property('event.ce-time');

    return response;
  }, retriesLeft, 2 * 1000)
      .catch((err) => {
        throw convertAxiosError(err, 'Fetching published legacy event responded with error');
      });
}

function getRandomEventId(encoding) {
  return 'event-' + encoding + '-' + genRandom(5);
}

async function getVirtualServiceHost(targetNamespace, funcName) {
  const vs = await waitForVirtualService(targetNamespace, funcName);
  return vs.spec.hosts[0];
}

async function checkInClusterEventDeliveryHelper(targetNamespace, encoding) {
  const eventId = getRandomEventId(encoding);
  const mockHost = await getVirtualServiceHost(targetNamespace, 'lastorder');

  if (isDebugEnabled()) {
    await printStatusOfInClusterEventingInfrastructure(targetNamespace, encoding, 'lastorder');
  }

  await sendInClusterEventWithRetry(mockHost, eventId, encoding);
  return ensureInClusterEventReceivedWithRetry(mockHost, eventId);
}

async function checkInClusterLegacyEvent(targetNamespace) {
  const eventId = getRandomEventId('legacy');
  const mockHost = await getVirtualServiceHost(targetNamespace, 'lastorder');

  if (isDebugEnabled()) {
    await printStatusOfInClusterEventingInfrastructure(targetNamespace, 'legacy', 'lastorder');
  }

  const eventData = {'id': eventId, 'legacyOrder': '987'};
  await sendInClusterLegacyEventWithRetry(mockHost, eventData);
  return ensureInClusterLegacyEventReceivedWithRetry(mockHost, eventId);
}

module.exports = {
  ensureCommerceMockLocalTestFixture,
  ensureCommerceMockWithCompassTestFixture,
  sendLegacyEventAndCheckResponse,
  sendCloudEventStructuredModeAndCheckResponse,
  sendCloudEventBinaryModeAndCheckResponse,
  sendLegacyEventAndCheckTracing,
  sendCloudEventStructuredModeAndCheckTracing,
  sendCloudEventBinaryModeAndCheckTracing,
  addService,
  updateService,
  deleteService,
  checkFunctionResponse,
  checkInClusterEventDelivery,
  checkInClusterEventTracing,
  cleanMockTestFixture,
  deleteMockTestFixture,
  waitForSubscriptionsTillReady,
  waitForSubscriptions,
  setEventMeshSourceNamespace,
  cleanCompassResourcesSKR,
  sendEventAndCheckResponse,
  getRandomEventId,
  getVirtualServiceHost,
  sendInClusterEventWithRetry,
  ensureInClusterEventReceivedWithRetry,
};
