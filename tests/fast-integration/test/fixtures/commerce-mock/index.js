const k8s = require("@kubernetes/client-node");
const fs = require("fs");
const path = require("path");
const { expect, config } = require("chai");
const https = require("https");
const axios = require("axios").default;
const httpsAgent = new https.Agent({
  rejectUnauthorized: false, // curl -k
});
axios.defaults.httpsAgent = httpsAgent;

const {
  retryPromise,
  expectNoK8sErr,
  convertAxiosError,
  sleep,
  k8sApply,
  waitForK8sObject,
  waitForServiceClass,
  waitForServiceInstance,
  waitForServiceBinding,
  waitForServiceBindingUsage,
  waitForVirtualService,
  waitForDeployment,
  waitForTokenRequest,
  waitForCompassConnection,
  deleteAllK8sResources,
  k8sAppsApi,
  k8sDynamicApi,
  deleteNamespaces,
  debug,
  toBase64,
} = require("../../../utils");

const {
  removeScenarioFromCompass,
  addScenarioInCompass,
  queryRuntimesForScenario,
  queryApplicationsForScenario,
  registerOrReturnApplication
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

const commerceObjs = k8s.loadAllYaml(commerceMockYaml);
const applicationObjs = k8s.loadAllYaml(applicationMockYaml);
const lastorderObjs = k8s.loadAllYaml(lastorderFunctionYaml);

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
    metadata: { name },
    spec: { serviceClassExternalName },
  };
}

async function checkAppGatewayResponse() {
  const vs = await waitForVirtualService("mocks", "commerce-mock");
  const mockHost = vs.spec.hosts[0];
  const host = mockHost.split(".").slice(1).join(".");
  let res = await retryPromise(
    () => axios.post(`https://lastorder.${host}`, { orderCode: "789" },{timeout:5000}),
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
        .get(`https://lastorder.${host}`)
        .then((res) => {
          expect(res.data).to.have.nested.property(
            "event.ce-type",
            "order.created"
          );
          expect(res.data).to.have.nested.property(
            "event.ce-source",
            "commerce"
          );
          expect(res.data).to.have.nested.property(
            "event.ce-eventtypeversion",
            "v1"
          );
          expect(res.data).to.have.nested.property(
            "event.ce-specversion",
            "1.0"
          );
          expect(res.data).to.have.nested.property("event.ce-id");
          expect(res.data).to.have.nested.property("event.ce-time");
          return res;
        })
        .catch((e) => { throw convertAxiosError(e, "Error during request to function lastorder") });
    },
    45,
    2 * 1000
  );
}

async function registerAllApis(mockHost) {
  const localApis = await retryPromise(
    () => axios.get(`https://${mockHost}/local/apis`, { timeout: 5000 }).catch((err) => {
      throw convertAxiosError(err, "API registration error - commerce mock local API not available");
    }
    ),
    1,
    3000
  );
  const filteredApis = localApis.data.filter((api) => (api.name.includes("Commerce Webservices") || api.name.includes("Events")));
  for (let api of filteredApis) {
    await retryPromise(
      async () => {
        await axios
          .post(
            `https://${mockHost}/local/apis/${api.id}/register`,
            {},
            {
              headers: {
                "content-type": "application/json",
                origin: `https://${mockHost}`,
              },
              timeout: 5000
            }
          ).catch((err) => {
            throw convertAxiosError(err, "Error during Commerce Mock API registration");
          });
      },
      10,
      3000
    );
  }

  const remoteApis = await axios
    .get(`https://${mockHost}/remote/apis`)
    .catch((err) => {
      throw convertAxiosError(err, "Commerce Mock registered apis not available");
    });
  expect(remoteApis.data).to.have.lengthOf.at.least(2);
  return remoteApis;
}

function connectMock(targetNamespace) {
  return async function(mockHost) {
    await k8sApply(applicationObjs);

    const tokenRequest = {
      apiVersion: "applicationconnector.kyma-project.io/v1alpha1",
      kind: "TokenRequest",
      metadata: { name: "commerce", namespace: targetNamespace },
    };

    await k8sDynamicApi.delete(tokenRequest).catch(() => {}); // Ignore delete error
    await k8sDynamicApi.create(tokenRequest);
    const tokenObj = await waitForTokenRequest("commerce", targetNamespace);
    
    const pairingBody = {
      token: tokenObj.status.url,
      baseUrl: `https://${mockHost}`,
      insecure: true,
    };
    
    await connectCommerceMock(mockHost, pairingBody);
    await ensureApplicationMapping("commerce", targetNamespace);
  }
}

function connectMockCompass(client, appName, scenarioName, targetNamespace) {
  return async function (mockHost) {
    const appID = await registerOrReturnApplication(client, appName, scenarioName);
    debug(`Application ID in Compass ${appID}`);

    const pairingData = await client.requestOneTimeTokenForApplication(appID);
    const pairingToken = toBase64(JSON.stringify(pairingData));
    const pairingBody = {
      token: pairingToken,
      baseUrl: mockHost,
      insecure: false
    };
    
    await connectCommerceMock(mockHost, pairingBody);
    await ensureApplicationMapping(`mp-${appName}`, targetNamespace);
  }
}

async function ensureApplicationMapping(name, ns) {
  const applicationMapping = {
    apiVersion: "applicationconnector.kyma-project.io/v1alpha1",
    kind: "ApplicationMapping",
    metadata: { name: name, namespace: ns }
  }
  await k8sDynamicApi.delete(applicationMapping).catch(() => {}); // Ignore delete error
  return await k8sDynamicApi.create(applicationMapping);
}

async function connectCommerceMock(mockHost, tokenData) {
  const url = `https://${mockHost}/connection`;
  const body = tokenData;
  const params = {
    headers: {
      "Content-Type": "application/json"
    }
  };

  try {
    await axios.post(url, body, params);
  } catch(err) {
    throw new Error(`Error during establishing connection from Commerce Mock to Kyma connector service: ${err.response.data}`);
  }
}

async function patchAppGatewayDeployment() {
  const commerceApplicationGatewayDeployment = await retryPromise(
    async () => {
      return k8sAppsApi.readNamespacedDeployment(
        "commerce-application-gateway",
        "kyma-integration"
      );
    },
    12,
    5000
  ).catch((err) => {
    throw new Error("Timeout: commerce-application-gateway is not ready");
  });
  expect(
    commerceApplicationGatewayDeployment.body.spec.template.spec.containers[0]
      .args[6]
  ).to.match(/^--skipVerify/);
  const patch = [
    {
      op: "replace",
      path: "/spec/template/spec/containers/0/args/6",
      value: "--skipVerify=true",
    },
  ];
  const options = {
    headers: { "Content-type": k8s.PatchUtils.PATCH_FORMAT_JSON_PATCH },
  };
  await k8sDynamicApi
    .requestPromise({
      url:
        k8sDynamicApi.basePath +
        commerceApplicationGatewayDeployment.body.metadata.selfLink,
      method: "PATCH",
      body: patch,
      json: true,
      headers: options.headers,
    })
    .catch(expectNoK8sErr);

  const patchedDeployment = await k8sAppsApi.readNamespacedDeployment(
    "commerce-application-gateway",
    "kyma-integration"
  );
  expect(
    patchedDeployment.body.spec.template.spec.containers[0].args[6]
  ).to.equal("--skipVerify=true");
  return patchedDeployment;
}

async function ensureCommerceMockTestFixture(mockNamespace, targetNamespace, connectFn) {
  await k8sApply([namespaceObj(mockNamespace), namespaceObj(targetNamespace)]);
  await k8sApply(commerceObjs);
  await k8sApply(lastorderObjs, targetNamespace, true);
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

  await retryPromise(() => connectFn(mockHost), 10, 3000);
  await retryPromise(() => registerAllApis(mockHost), 10, 3000);

  const commerceSC = await waitForServiceClass("commerce", targetNamespace, 300 * 1000);
  const commerceSCExternalName = commerceSC.spec.externalName;
  const serviceCatalogObjs = [
    serviceInstanceObj("commerce", commerceSCExternalName),
  ];

  await retryPromise(
    () => k8sApply(serviceCatalogObjs, targetNamespace, false),
    5,
    2000
  );
  await waitForServiceInstance("commerce", targetNamespace);
  
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

  await patchAppGatewayDeployment();

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

async function registerKymaInCompass(client, runtimeName, scenarioName) {
  await addScenarioInCompass(client, scenarioName);
  const runtimeID = await client.registerRuntime(runtimeName, scenarioName);
  debug(`Runtime ID in Compass ${runtimeID}`);
  
  const pairingData = await client.requestOneTimeTokenForRuntime(runtimeID);
  const compassAgentCfg = {
    apiVersion: "v1",
    kind: "Secret",
    metadata: {
      name: "compass-agent-configuration",
    },
    data: {
      CONNECTOR_URL: toBase64(pairingData.connectorURL),
      RUNTIME_ID: toBase64(runtimeID),
      TENANT: toBase64(client.tenantID),
      TOKEN: toBase64(pairingData.token),
    }
  };
  await k8sApply([compassAgentCfg], "compass-system");
  await waitForCompassConnection("compass-connection");
}

async function unregisterKymaFromCompass(client, scenarioName) {
  // Cleanup Compass
  const applications = await queryApplicationsForScenario(client, scenarioName);
  for(let application of applications) {
    await client.unregisterApplication(application.id);
  }

  // TODO: refactor this step to cover runtime agent deleting the application from Kyma
  // and then remove the runtime from compass

  // Delete connection between Compass Agent and Compass
  deleteAllK8sResources("/api/v1/namespaces/compass-system/secrets/compass-agent-configuration");
  deleteAllK8sResources("/apis/compass.kyma-project.io/v1alpha1/compassconnections/compass-connection");

  const runtimes = await queryRuntimesForScenario(client, scenarioName);
  for(let runtime of runtimes) {
    await client.unregisterRuntime(runtime.id);
  }
  
  await removeScenarioFromCompass(client, scenarioName);
}

module.exports = {
  ensureCommerceMockTestFixture,
  sendEventAndCheckResponse,
  checkAppGatewayResponse,
  cleanMockTestFixture,
  connectMock,
  connectMockCompass,
  registerKymaInCompass,
  unregisterKymaFromCompass,
};
