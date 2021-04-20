const k8s = require("@kubernetes/client-node");
const fs = require("fs");
const path = require("path");
const { expect } = require("chai");
const https = require("https");
const axios = require("axios").default;
const httpsAgent = new https.Agent({
  rejectUnauthorized: false, // curl -k
});
axios.defaults.httpsAgent = httpsAgent;

const {
  retryPromise,
  convertAxiosError,
  sleep,
  k8sApply,
  waitForServiceClass,
  waitForServiceInstance,
  waitForServiceBinding,
  waitForServiceBindingUsage,
  waitForVirtualService,
  waitForDeployment,
  waitForTokenRequest,
  deleteAllK8sResources,
  k8sDynamicApi,
  deleteNamespaces,
  debug,
  toBase64,
  ensureApplicationMapping,
  patchApplicationGateway
} = require("../../../utils");

const {
  registerOrReturnApplication,
} = require("../../../compass");

const commerceMockYaml = fs.readFileSync(
  path.join(__dirname, "./commerce-mock.yaml"),
  {
    encoding: "utf8",
  }
);

const applicationMockYaml = fs.readFileSync(
  path.join(__dirname, "./application.yaml"),
  {
    encoding: "utf8",
  }
);


const lastorderFunctionYaml = fs.readFileSync(
  path.join(__dirname, "./lastorder-function.yaml"),
  {
    encoding: "utf8",
  }
);

const lastorderCentralApplicationGatewayFunctionYaml = fs.readFileSync(
    path.join(__dirname, "..", "..", "..", "application-connectivity-2-test", "./lastorder-function.yaml"),
    {
      encoding: "utf8",
    }
);

const lastorderCentralApplicationGatewayAndCompassFunctionYaml = fs.readFileSync(
    path.join(__dirname, "..", "..", "..", "application-connectivity-2-compass-test", "./lastorder-function.yaml"),
    {
      encoding: "utf8",
    }
);

const commerceObjs = k8s.loadAllYaml(commerceMockYaml);
const applicationObjs = k8s.loadAllYaml(applicationMockYaml);
const lastorderObjs = k8s.loadAllYaml(lastorderFunctionYaml);
const lastorderCentralApplicationGatewayObjs = k8s.loadAllYaml(lastorderCentralApplicationGatewayFunctionYaml);
const lastorderCentralApplicationGatewayAndCompassObjs = k8s.loadAllYaml(lastorderCentralApplicationGatewayAndCompassFunctionYaml);

function namespaceObj(name) {
  return {
    apiVersion: "v1",
    kind: "Namespace",
    metadata: { name },
  };
}

function serviceInstanceObj(name, serviceClassExternalName) {
  return {
    apiVersion: "servicecatalog.k8s.io/v1beta1",
    kind: "ServiceInstance",
    metadata: {
      name: name,
    },
    spec: { serviceClassExternalName },
  };
}


function eventingSubscription(eventType, sink, fnName, ns) {
  return {
    apiVersion: "eventing.kyma-project.io/v1alpha1",
    kind: "Subscription",
    metadata: {
      name: `function-${fnName}`,
      namespace: ns,
    },
    spec: {
      filter: {
        dialect: "beb",
        filters: [{
          eventSource: {
            property: "source", type: "exact", value: "",
          },
          eventType: {
            property: "type",type: "exact", value: eventType/*sap.kyma.custom.commerce.order.created.v1*/
          } 
        }]
      },
      protocol: "BEB",
      protocolsettings: {
        exemptHandshake: true,
        qos: "AT-LEAST-ONCE",
      },
      sink: sink/*http://lastorder.test.svc.cluster.local*/
    }
  }
}

async function checkAppGatewayResponse() {
  const vs = await waitForVirtualService("mocks", "commerce-mock");
  const mockHost = vs.spec.hosts[0];
  const host = mockHost.split(".").slice(1).join(".");
  
  let res = await retryPromise(
    () => axios.post(`https://lastorder.${host}`, { orderCode: "789" }, { timeout: 5000 }),
    45,
    2000
  ).catch((err) => { throw convertAxiosError(err, "Function lastorder responded with error") });
  
  expect(res.data).to.have.nested.property(
    "order.totalPriceWithTax.value",
    100
  );
}

async function sendEventAndCheckResponse() {
  const vs = await waitForVirtualService("mocks", "commerce-mock");
  const mockHost = vs.spec.hosts[0];
  const host = mockHost.split(".").slice(1).join(".");

  await retryPromise(
    async () => {
      await axios
        .post(
          `https://${mockHost}/events`,
          {
            "event-type": "order.created",
            "event-type-version": "v1",
            "event-time": "2020-09-28T14:47:16.491Z",
            data: { orderCode: "567" },
            "event-tracing": true,
          },
          {
            headers: {
              "content-type": "application/json",
            },
          }
        )
        .catch((e) => {
          console.log("Cannot send event, the response from event gateway:");
          console.dir(e.response.data);
          throw convertAxiosError(e, "Cannot send event, the response from event gateway");
        });

      await sleep(500);

      return axios
        .get(`https://lastorder.${host}`, { timeout: 5000 })
        .then((res) => {
          expect(res.data).to.have.nested.property("event.data.orderCode", "567");
          // See: https://github.com/kyma-project/kyma/issues/10720
          expect(res.data).to.have.nested.property("event.ce-type").that.contains("order.created");
          expect(res.data).to.have.nested.property("event.ce-source");
          expect(res.data).to.have.nested.property("event.ce-eventtypeversion", "v1");
          expect(res.data).to.have.nested.property("event.ce-specversion", "1.0");
          expect(res.data).to.have.nested.property("event.ce-id");
          expect(res.data).to.have.nested.property("event.ce-time");
          return res;
        })
        .catch((e) => {
          throw convertAxiosError(e, "Error during request to function lastorder")
        });
    },
    30,
    2 * 1000
  );
}

async function registerAllApis(mockHost) {
  debug("Listing Commerce Mock local APIs")
  const localApis = await axios.get(`https://${mockHost}/local/apis`, { timeout: 5000 }).catch((err) => {
    throw convertAxiosError(err, "API registration error - commerce mock local API not available");
  });
  debug("Commerce Mock local APIs received")
  const filteredApis = localApis.data.filter((api) => (api.name.includes("Commerce Webservices") || api.name.includes("Events")));
  for (let api of filteredApis) {
    debug("Registering", api.name)
    await axios
      .post(
        `https://${mockHost}/local/apis/${api.id}/register`,
        {},
        {
          headers: {
            "content-type": "application/json",
            origin: `https://${mockHost}`,
          },
          timeout: 30000
        }
      ).catch((err) => {
        throw convertAxiosError(err, "Error during Commerce Mock API registration");
      });
  }
  debug("Verifying if APIs are properly registered")

  const remoteApis = await axios
    .get(`https://${mockHost}/remote/apis`)
    .catch((err) => {
      throw convertAxiosError(err, "Commerce Mock registered apis not available");
    });
  expect(remoteApis.data).to.have.lengthOf.at.least(2);
  debug("Commerce APIs registered");
  return remoteApis;
}

async function connectMockLocal(mockHost, targetNamespace) {
  const tokenRequest = {
    apiVersion: "applicationconnector.kyma-project.io/v1alpha1",
    kind: "TokenRequest",
    metadata: { name: "commerce", namespace: targetNamespace },
  };
  await k8sDynamicApi.delete(tokenRequest).catch(() => { }); // Ignore delete error
  await k8sDynamicApi.create(tokenRequest);
  const tokenObj = await waitForTokenRequest("commerce", targetNamespace);

  const pairingBody = {
    token: tokenObj.status.url,
    baseUrl: `https://${mockHost}`,
    insecure: true,
  };
  debug("Token URL", tokenObj.status.url);
  await connectCommerceMock(mockHost, pairingBody);
  await ensureApplicationMapping("commerce", targetNamespace);
  debug("Commerce mock connected locally");
}

async function connectMockCompass(client, appName, scenarioName, mockHost, targetNamespace) {
  const appID = await registerOrReturnApplication(client, appName, scenarioName);
  debug(`Application ID in Compass ${appID}`);

  const pairingData = await client.requestOneTimeTokenForApplication(appID);
  const pairingToken = toBase64(JSON.stringify(pairingData));
  const pairingBody = {
    token: pairingToken,
    baseUrl: mockHost,
    insecure: true
  };
  
  debug(`Connecting ${mockHost}`);
  await connectCommerceMock(mockHost, pairingBody);
  
  debug(`Creating application mapping for mp-${appName} in ${targetNamespace}`);
  await ensureApplicationMapping(`mp-${appName}`, targetNamespace);
  debug("Commerce mock connected to Compass");
}

async function connectCommerceMock(mockHost, tokenData) {
  const url = `https://${mockHost}/connection`;
  const body = tokenData;
  const params = {
    headers: {
      "Content-Type": "application/json"
    },
    timeout: 5000,
  };

  try {
    await axios.post(url, body, params);
  } catch (err) {
    throw convertAxiosError(err, "Error during establishing connection from Commerce Mock to Kyma connector service");
  }
}

async function ensureCommerceMockWithCompassTestFixture(client, appName, scenarioName, mockNamespace, targetNamespace, withCentralApplicationGateway=false) {
  const mockHost = await provisionCommerceMockResources(
      `mp-${appName}`,
      mockNamespace,
      targetNamespace,
      withCentralApplicationGateway ? lastorderCentralApplicationGatewayAndCompassObjs : lastorderObjs);
  await retryPromise(() => connectMockCompass(client, appName, scenarioName, mockHost, targetNamespace), 10, 3000);
  await retryPromise(() => registerAllApis(mockHost), 10, 3000);
  await waitForDeployment(`mp-${appName}-connectivity-validator`, "kyma-integration");

  const commerceSC = await waitForServiceClass(appName, targetNamespace, 300 * 1000);
  
  await retryPromise(
    () => k8sApply([serviceInstanceObj("commerce", commerceSC.spec.externalName)], targetNamespace, false),
    5,
    2000
  );
  await waitForServiceInstance("commerce", targetNamespace, 300 * 1000);

  await patchApplicationGateway(`${targetNamespace}-gateway`, targetNamespace);

  const serviceBinding = {
    apiVersion: "servicecatalog.k8s.io/v1beta1",
    kind: "ServiceBinding",
    metadata: { name: "commerce-binding" },
    spec: {
      instanceRef: { name: "commerce" },
    },
  };
  await k8sApply([serviceBinding], targetNamespace, false);
  await waitForServiceBinding("commerce-binding", targetNamespace);

  const serviceBindingUsage = {
    apiVersion: "servicecatalog.kyma-project.io/v1alpha1",
    kind: "ServiceBindingUsage",
    metadata: { name: "commerce-lastorder-sbu" },
    spec: {
      serviceBindingRef: { name: "commerce-binding" },
      usedBy: { kind: "serverless-function", name: "lastorder" },
    },
  };
  await k8sApply([serviceBindingUsage], targetNamespace);
  await waitForServiceBindingUsage("commerce-lastorder-sbu", targetNamespace);

  return mockHost;
}

async function ensureCommerceMockLocalTestFixture(mockNamespace, targetNamespace, withCentralApplicationGateway=false) {
  
  await k8sApply(applicationObjs);
  const mockHost = await provisionCommerceMockResources(
      "commerce",
      mockNamespace,
      targetNamespace,
      withCentralApplicationGateway ? lastorderCentralApplicationGatewayObjs : lastorderObjs);
  await retryPromise(() => connectMockLocal(mockHost, targetNamespace), 10, 3000);
  await retryPromise(() => registerAllApis(mockHost), 10, 3000);

  const webServicesSC = await waitForServiceClass(
    "webservices",
    targetNamespace,
    300 * 1000
  );
  const eventsSC = await waitForServiceClass("events", targetNamespace);
  const webServicesSCExternalName = webServicesSC.spec.externalName;
  const eventsSCExternalName = eventsSC.spec.externalName;
  const serviceCatalogObjs = [
    serviceInstanceObj("commerce-webservices", webServicesSCExternalName),
    serviceInstanceObj("commerce-events", eventsSCExternalName),
  ];

  await retryPromise(
    () => k8sApply(serviceCatalogObjs, targetNamespace, false),
    5,
    2000
  );
  await waitForServiceInstance("commerce-webservices", targetNamespace);
  await waitForServiceInstance("commerce-events", targetNamespace);

  const serviceBinding = {
    apiVersion: "servicecatalog.k8s.io/v1beta1",
    kind: "ServiceBinding",
    metadata: {
      name: "commerce-binding",
    },
    spec: {
      instanceRef: { name: "commerce-webservices" },
    },
  };
  await k8sApply([serviceBinding], targetNamespace, false);
  await waitForServiceBinding("commerce-binding", targetNamespace);

  const serviceBindingUsage = {
    apiVersion: "servicecatalog.kyma-project.io/v1alpha1",
    kind: "ServiceBindingUsage",
    metadata: { name: "commerce-lastorder-sbu" },
    spec: {
      serviceBindingRef: { name: "commerce-binding" },
      usedBy: { kind: "serverless-function", name: "lastorder" },
    },
  };
  await k8sApply([serviceBindingUsage], targetNamespace);
  await waitForServiceBindingUsage("commerce-lastorder-sbu", targetNamespace);

  return mockHost;
}

async function provisionCommerceMockResources(appName, mockNamespace, targetNamespace, functionObjs=lastorderObjs) {
  await k8sApply([namespaceObj(mockNamespace), namespaceObj(targetNamespace)]);
  await k8sApply(commerceObjs);
  await k8sApply(functionObjs, targetNamespace, true);
  await k8sApply([
    eventingSubscription(
      `sap.kyma.custom.${appName}.order.created.v1`,
      `http://lastorder.${targetNamespace}.svc.cluster.local`,
      "lastorder",
      targetNamespace)
  ]);
  await waitForDeployment("commerce-mock", "mocks", 120 * 1000);
  const vs = await waitForVirtualService("mocks", "commerce-mock");
  const mockHost = vs.spec.hosts[0];
  await retryPromise(
    () =>
      axios.get(`https://${mockHost}/local/apis`).catch((err) => {
        throw convertAxiosError(err, "Commerce mock local API not available - timeout");
      }),
    40,
    3000
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

function cleanMockTestFixture(mockNamespace, targetNamespace, wait = true) {
  for (let path of getResourcePaths(mockNamespace).concat(
    getResourcePaths(targetNamespace)
  )) {
    deleteAllK8sResources(path);
  }
  k8sDynamicApi.delete({
    apiVersion: "applicationconnector.kyma-project.io/v1alpha1",
    kind: "Application",
    metadata: {
      name: "commerce",
    },
  });
  return deleteNamespaces([mockNamespace, targetNamespace], wait);
}
module.exports = {
  ensureCommerceMockLocalTestFixture,
  ensureCommerceMockWithCompassTestFixture,
  sendEventAndCheckResponse,
  checkAppGatewayResponse,
  cleanMockTestFixture,
};
