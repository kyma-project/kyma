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
  ignore404,
  sleep,
  k8sApply,
  waitForServiceClass,
  waitForServicePlanByServiceClass,
  waitForServiceInstance,
  waitForServiceBinding,
  waitForServiceBindingUsage,
  waitForVirtualService,
  waitForDeployment,
  waitForTokenRequest,
  waitForFunction,
  waitForSubscription,
  deleteAllK8sResources,
  genRandom,
  k8sDynamicApi,
  deleteNamespaces,
  debug,
  toBase64,
  ensureApplicationMapping,
  patchApplicationGateway,
  eventingSubscription,
  k8sDelete,
  getSecretData,
  namespaceObj,
  serviceInstanceObj,
  getTraceDAG,
  printStatusOfInClusterEventingInfrastructure,
} = require('../../../utils');

const {
  registerOrReturnApplication,
  deregisterApplication,
  removeApplicationFromScenario,
  removeScenarioFromCompass,
  getApplicationByName,
  unassignRuntimeFromScenario,
} = require('../../../compass');

const {
  jaegerPortForward,
  getJaegerTrace,
} = require('../../../monitoring/client');

const {
  OAuthToken,
  OAuthCredentials,
} = require('../../../lib/oauth');

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
      const orderWithCompass = `${gatewayUrl}:8082/%%APP_NAME%%/sap-commerce-cloud/commerce-webservices/site/orders/`;
      return k8s.loadAllYaml(functionYaml.toString()
          .replace('%%URL%%', `"${orderWithCompass}" + code`)
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
  let res = await retryPromise(
      () => axios.post(`https://lastorder.${host}/function`, {orderCode: '789'}, {
        timeout: 5000,
        headers: {Authorization: `bearer ${accessToken}`},
      }),
      45,
      2000,
  ).catch((err) => {
    throw convertAxiosError(err, 'Function lastorder responded with error');
  });

  expect(res.data).to.have.nested.property('order.totalPriceWithTax.value', 100);

  // expect error when unauthorized
  let errorOccurred = false;
  try {
    res = await axios.post(`https://lastorder.${host}/function`, {orderCode: '789'}, {timeout: 5000});
  } catch (err) {
    errorOccurred = true;
    expect(err.response.status).to.be.equal(401);
  }
  expect(errorOccurred).to.be.equal(true);
}

async function sendEventAndCheckResponse(mockNamespace = 'mocks') {
  const vs = await waitForVirtualService(mockNamespace, 'commerce-mock');
  const mockHost = vs.spec.hosts[0];
  const host = mockHost.split('.').slice(1).join('.');

  return await retryPromise(
      async () => {
        await axios
            .post(
                `https://${mockHost}/events`,
                {
                  'event-type': 'order.created',
                  'event-type-version': 'v1',
                  'event-time': '2020-09-28T14:47:16.491Z',
                  'data': {'orderCode': '567'},
                  'event-tracing': true,
                },
                {
                  headers: {
                    'content-type': 'application/json',
                  },
                },
            )
            .catch((e) => {
              console.log('Cannot send event, the response from event gateway:');
              console.dir(e.response.data);
              throw convertAxiosError(e, 'Cannot send event, the response from event gateway');
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
      30,
      2 * 1000,
  );
}

async function sendLegacyEventAndCheckTracing(targetNamespace = 'test', mockNamespace = 'mocks') {
  // Send an event and get it back from the lastorder function
  const res = await sendEventAndCheckResponse(mockNamespace);
  expect(res.data).to.have.nested.property('event.headers.x-b3-traceid');
  expect(res.data).to.have.nested.property('podName');

  // Extract traceId from response
  const traceId = res.data.event.headers['x-b3-traceid'];

  // Define expected trace data
  const correctTraceSpansLength = 6;
  const correctTraceProcessSequence = [
    'istio-ingressgateway.istio-system',
    'central-application-connectivity-validator.kyma-system',
    'central-application-connectivity-validator.kyma-system',
    'eventing-publisher-proxy.kyma-system',
    'eventing-controller.kyma-system',
    `lastorder-${res.data.podName.split('-')[1]}.${targetNamespace}`,
  ];

  // wait sometime for jaeger to complete tracing data
  await sleep(10 * 1000);
  await checkTrace(traceId, correctTraceSpansLength, correctTraceProcessSequence);
}

async function checkInClusterEventTracing(targetNamespace) {
  const res = await checkInClusterEventDeliveryHelper(targetNamespace, 'structured');
  expect(res.data).to.have.nested.property('event.headers.x-b3-traceid');
  expect(res.data).to.have.nested.property('podName');

  // Extract traceId from response
  const traceId = res.data.event.headers['x-b3-traceid'];

  // Define expected trace data
  const correctTraceSpansLength = 4;
  const correctTraceProcessSequence = [
    // We are sending the in-cluster event from inside the lastorder pod
    `lastorder-${res.data.podName.split('-')[1]}.${targetNamespace}`,
    'eventing-publisher-proxy.kyma-system',
    'eventing-controller.kyma-system',
    `lastorder-${res.data.podName.split('-')[1]}.${targetNamespace}`,
  ];

  // wait sometime for jaeger to complete tracing data
  await sleep(10 * 1000);
  await checkTrace(traceId, correctTraceSpansLength, correctTraceProcessSequence);
}

async function checkTrace(traceId, expectedTraceLength, expectedTraceProcessSequence) {
  // Port-forward to Jaeger and fetch trace data for the traceId
  const cancelJaegerPortForward = await jaegerPortForward();
  let traceRes;
  try {
    traceRes = await getJaegerTrace(traceId);
  } finally {
    cancelJaegerPortForward();
  }

  // the trace response should have data for single trace
  expect(traceRes.data).to.have.length(1);

  // Extract trace data from response
  const traceData = traceRes.data[0];
  expect(traceData['spans']).to.have.length(expectedTraceLength);

  // Generate DAG for trace spans
  const traceDAG = await getTraceDAG(traceData);
  expect(traceDAG).to.have.length(1);

  // Check the tracing spans are correct
  let currentSpan = traceDAG[0];
  for (let i = 0; i < expectedTraceLength; i++) {
    const processServiceName = traceData.processes[currentSpan.processID].serviceName;
    debug(`Checking Trace Sequence # ${i}:
    Expected process: ${expectedTraceProcessSequence[i]}, 
    Received process: ${processServiceName}`);
    expect(processServiceName).to.be.equal(expectedTraceProcessSequence[i]);

    // Traverse to next trace span
    if (i < expectedTraceLength - 1) {
      expect(currentSpan.childSpans).to.have.length(1);
      currentSpan = currentSpan.childSpans[0];
    }
  }
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

async function connectMockLocal(mockHost, targetNamespace) {
  const tokenRequest = {
    apiVersion: 'applicationconnector.kyma-project.io/v1alpha1',
    kind: 'TokenRequest',
    metadata: {name: 'commerce', namespace: targetNamespace},
  };
  await k8sDynamicApi.delete(tokenRequest).catch(() => { }); // Ignore delete error
  await k8sDynamicApi.create(tokenRequest);
  const tokenObj = await waitForTokenRequest('commerce', targetNamespace);

  const pairingBody = {
    token: tokenObj.status.url,
    baseUrl: `https://${mockHost}`,
    insecure: true,
  };
  debug('Token URL', tokenObj.status.url);
  await connectCommerceMock(mockHost, pairingBody);
  await ensureApplicationMapping('commerce', targetNamespace);
  debug('Commerce mock connected locally');
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

  debug(`Creating application mapping for mp-${appName} in ${targetNamespace}`);
  await ensureApplicationMapping(`mp-${appName}`, targetNamespace);
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

async function ensureCommerceMockWithCompassTestFixture(client,
    appName,
    scenarioName,
    mockNamespace,
    targetNamespace,
    withCentralApplicationConnectivity = false) {
  const lastOrderFunction = withCentralApplicationConnectivity ?
    prepareFunction('central-app-gateway-compass', `mp-${appName}`) :
    prepareFunction();

  const mockHost = await provisionCommerceMockResources(
      `mp-${appName}`,
      mockNamespace,
      targetNamespace,
      lastOrderFunction);
  await retryPromise(() => connectMockCompass(client, appName, scenarioName, mockHost, targetNamespace), 10, 3000);
  await retryPromise(() => registerAllApis(mockHost), 10, 3000);

  const commerceSC = await waitForServiceClass(appName, targetNamespace, 300 * 1000);
  await waitForServicePlanByServiceClass(commerceSC.metadata.name, targetNamespace, 300 * 1000);
  await retryPromise(
      () => k8sApply([serviceInstanceObj('commerce', commerceSC.spec.externalName)], targetNamespace, false),
      5,
      2000,
  );
  await waitForServiceInstance('commerce', targetNamespace, 600 * 1000);

  if (withCentralApplicationConnectivity) {
    await waitForDeployment('central-application-gateway', 'kyma-system');
    await waitForDeployment('central-application-connectivity-validator', 'kyma-system');
    await patchApplicationGateway('central-application-gateway', 'kyma-system');
  } else {
    await waitForDeployment(`${targetNamespace}-gateway`, targetNamespace);
    await patchApplicationGateway(`${targetNamespace}-gateway`, targetNamespace);
  }

  const serviceBinding = {
    apiVersion: 'servicecatalog.k8s.io/v1beta1',
    kind: 'ServiceBinding',
    metadata: {name: 'commerce-binding'},
    spec: {
      instanceRef: {name: 'commerce'},
    },
  };
  await k8sApply([serviceBinding], targetNamespace, false);
  await waitForServiceBinding('commerce-binding', targetNamespace);

  const serviceBindingUsage = {
    apiVersion: 'servicecatalog.kyma-project.io/v1alpha1',
    kind: 'ServiceBindingUsage',
    metadata: {name: 'commerce-lastorder-sbu'},
    spec: {
      serviceBindingRef: {name: 'commerce-binding'},
      usedBy: {kind: 'serverless-function', name: 'lastorder'},
    },
  };
  await k8sApply([serviceBindingUsage], targetNamespace);
  await waitForServiceBindingUsage('commerce-lastorder-sbu', targetNamespace);

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

async function ensureCommerceMockLocalTestFixture(mockNamespace,
    targetNamespace,
    withCentralApplicationConnectivity = false) {
  await k8sApply(applicationObjs);
  const mockHost = await provisionCommerceMockResources(
      'commerce',
      mockNamespace,
      targetNamespace,
    withCentralApplicationConnectivity ? prepareFunction('central-app-gateway') : prepareFunction());
  await retryPromise(() => connectMockLocal(mockHost, targetNamespace), 10, 3000);
  await retryPromise(() => registerAllApis(mockHost), 10, 3000);

  if (withCentralApplicationConnectivity) {
    await waitForDeployment('central-application-gateway', 'kyma-system');
    await waitForDeployment('central-application-connectivity-validator', 'kyma-system');
    await patchApplicationGateway('central-application-gateway', 'kyma-system');
  }

  const webServicesSC = await waitForServiceClass(
      'webservices',
      targetNamespace,
      400 * 1000,
  );
  const eventsSC = await waitForServiceClass('events', targetNamespace);
  const webServicesSCExternalName = webServicesSC.spec.externalName;
  const eventsSCExternalName = eventsSC.spec.externalName;
  const serviceCatalogObjs = [
    serviceInstanceObj('commerce-webservices', webServicesSCExternalName),
    serviceInstanceObj('commerce-events', eventsSCExternalName),
  ];

  await retryPromise(
      () => k8sApply(serviceCatalogObjs, targetNamespace, false),
      5,
      2000,
  );
  await waitForServiceInstance('commerce-webservices', targetNamespace);
  await waitForServiceInstance('commerce-events', targetNamespace);

  const serviceBinding = {
    apiVersion: 'servicecatalog.k8s.io/v1beta1',
    kind: 'ServiceBinding',
    metadata: {
      name: 'commerce-binding',
    },
    spec: {
      instanceRef: {name: 'commerce-webservices'},
    },
  };
  await k8sApply([serviceBinding], targetNamespace, false);
  await waitForServiceBinding('commerce-binding', targetNamespace);

  const serviceBindingUsage = {
    apiVersion: 'servicecatalog.kyma-project.io/v1alpha1',
    kind: 'ServiceBindingUsage',
    metadata: {name: 'commerce-lastorder-sbu'},
    spec: {
      serviceBindingRef: {name: 'commerce-binding'},
      usedBy: {kind: 'serverless-function', name: 'lastorder'},
    },
  };
  await k8sApply([serviceBindingUsage], targetNamespace);
  await waitForServiceBindingUsage('commerce-lastorder-sbu', targetNamespace);

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
    `/apis/servicecatalog.kyma-project.io/v1alpha1/namespaces/${namespace}/servicebindingusages`,
    `/apis/servicecatalog.k8s.io/v1beta1/namespaces/${namespace}/servicebindings`,
    `/apis/servicecatalog.k8s.io/v1beta1/namespaces/${namespace}/serviceinstances`,
    `/apis/serverless.kyma-project.io/v1alpha1/namespaces/${namespace}/functions`,
    `/apis/addons.kyma-project.io/v1alpha1/namespaces/${namespace}/addonsconfigurations`,
    `/apis/gateway.kyma-project.io/v1alpha1/namespaces/${namespace}/apirules`,
    `/apis/apps/v1/namespaces/${namespace}/deployments`,
    `/api/v1/namespaces/${namespace}/services`,
    `/apis/applicationconnector.kyma-project.io/v1alpha1/namespaces/${namespace}/applicationmappings`,
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
    ignore404(err);
  }

  debug('Deleting test namespaces');
  return deleteNamespaces([mockNamespace, targetNamespace], wait);
}

async function deleteMockTestFixture(mockNamespace) {
  const serviceBindingUsage = {
    apiVersion: 'servicecatalog.kyma-project.io/v1alpha1',
    kind: 'ServiceBindingUsage',
    metadata: {name: 'commerce-lastorder-sbu'},
    spec: {
      serviceBindingRef: {name: 'commerce-binding'},
      usedBy: {kind: 'serverless-function', name: 'lastorder'},
    },
  };
  await k8sDelete([serviceBindingUsage], mockNamespace);
  const serviceBinding = {
    apiVersion: 'servicecatalog.k8s.io/v1beta1',
    kind: 'ServiceBinding',
    metadata: {name: 'commerce-binding'},
    spec: {
      instanceRef: {name: 'commerce'},
    },
  };
  await k8sDelete([serviceBinding], mockNamespace, false);
  await k8sDelete(lastorderObjs);
  await k8sDelete(prepareCommerceObjs(mockNamespace));
  await k8sDelete(applicationObjs);
}

async function waitForSubscriptionsTillReady(targetNamespace) {
  await waitForSubscription('order-received', targetNamespace);
  await waitForSubscription('order-created', targetNamespace);
}

async function checkInClusterEventDelivery(targetNamespace) {
  await checkInClusterEventDeliveryHelper(targetNamespace, 'structured');
  await checkInClusterEventDeliveryHelper(targetNamespace, 'binary');
}

async function checkInClusterEventDeliveryHelper(targetNamespace, encoding) {
  const eventId = "event-" + encoding + "-" + genRandom(5);
  const vs = await waitForVirtualService(targetNamespace, "lastorder");
  const mockHost = vs.spec.hosts[0];

  await printStatusOfInClusterEventingInfrastructure(targetNamespace, encoding, "lastorder");

  // send event using function query parameter send=true
  await retryPromise(async () => {
    const response = await axios.post(`https://${mockHost}`, {id: eventId}, {
      params: {
        send: true,
        encoding: encoding
      }
    });
    debug("Event publishing result:", {
      status: response.status,
      statusText: response.statusText,
      data: response.data
    });
    if (response.data.eventPublishError) {
      throw convertAxiosError(response.data.statusText);
    }
    expect(response.status).to.be.equal(200);
  }, 10, 1000);
  debug(`Event ${eventId} was successfully published`);

  // verify if event was received using function query parameter inappevent=eventId
  return await retryPromise(async () => {
    debug('Waiting for event: ', eventId);
    const response = await axios.get(`https://${mockHost}`, {params: {inappevent: eventId}});
    debug("Received Event response: ", {
      status: response.status,
      statusText: response.statusText,
      data: response.data
    });
    expect(response.data).to.have.nested.property("event.id", eventId, "The same event id expected in the result");
    expect(response.data).to.have.nested.property("event.shipped", true, "Order should have property shipped");
    return response;
  }, 30, 2 * 1000)
      .catch((err) => {
        throw convertAxiosError(err, "Fetching published event responded with error");
      });
}

module.exports = {
  ensureCommerceMockLocalTestFixture,
  ensureCommerceMockWithCompassTestFixture,
  sendEventAndCheckResponse,
  sendLegacyEventAndCheckTracing,
  addService,
  updateService,
  deleteService,
  checkFunctionResponse,
  checkInClusterEventDelivery,
  checkInClusterEventTracing,
  cleanMockTestFixture,
  deleteMockTestFixture,
  waitForSubscriptionsTillReady,
  setEventMeshSourceNamespace,
  cleanCompassResourcesSKR,
};
