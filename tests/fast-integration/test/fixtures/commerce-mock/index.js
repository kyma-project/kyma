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
  expectNoAxiosErr,
  sleep,
  k8sApply,
  waitForK8sObject,
  waitForServiceClass,
  waitForServiceInstance,
  waitForServiceBinding,
  waitForServiceBindingUsage,
  waitForVirtualService,
  deleteAllK8sResources,
  k8sAppsApi,
  k8sDynamicApi,
  deleteNamespaces,
} = require('../../../utils');

const commerceMockYaml = fs.readFileSync(
  path.join(__dirname, "./commerce-mock.yaml"),
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
const lastorderObjs = k8s.loadAllYaml(lastorderFunctionYaml);

async function checkAppGatewayResponse() {
  const vs = await waitForVirtualService('mocks', 'commerce-mock')
  const mockHost = vs.spec.hosts[0]
  const host = mockHost.split(".").slice(1).join(".");
  let res = await retryPromise(() => axios.post(`https://lastorder.${host}`, { orderCode: "789" }), 10, 2000);
  expect(res.data).to.have.nested.property("order.totalPriceWithTax.value", 100);
}

async function sendEventAndCheckResponse() {
  const vs = await waitForVirtualService('mocks', 'commerce-mock')
  const mockHost = vs.spec.hosts[0]
  const host = mockHost.split(".").slice(1).join(".");

  await retryPromise(
    async () => {
      await axios.post(
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
      ).catch(e => {
        console.dir({ status: e.response.status, data: e.response.data });
        throw e
      });

      await sleep(500);

      return axios.get(`https://lastorder.${host}`).then((res) => {
        expect(res.data).to.have.nested.property("event.ce-type", 'order.created');
        expect(res.data).to.have.nested.property("event.ce-source", 'commerce');
        expect(res.data).to.have.nested.property("event.ce-eventtypeversion", 'v1');
        expect(res.data).to.have.nested.property("event.ce-specversion", "1.0");
        expect(res.data).to.have.nested.property("event.ce-id");
        expect(res.data).to.have.nested.property("event.ce-time");
        return res;
      });
    },
    30,
    2 * 1000
  ).catch(expectNoAxiosErr);
}

async function registerAllApis(mockHost) {
  const localApis = await axios.get(`https://${mockHost}/local/apis`).catch(expectNoAxiosErr);
  for (let api of localApis.data) {
    await retryPromise(async () => {
      await axios.post(`https://${mockHost}/local/apis/${api.id}/register`, {},
        {
          headers: {
            "content-type": "application/json",
            origin: `https://${mockHost}`,
          },
        }
      ).catch(expectNoAxiosErr)
    }, 3, 5000)
  }
  const remoteApis = await axios.get(`https://${mockHost}/remote/apis`).catch(expectNoAxiosErr);
  expect(remoteApis.data).to.have.lengthOf.at.least(2)
  return remoteApis;
}

function namespaceObj(name) {
  return {
    apiVersion: 'v1',
    kind: 'Namespace',
    metadata: { name }
  }

}

async function connectMock(mockHost, targetNamespace) {
  const tokenRequest = {
    apiVersion: 'applicationconnector.kyma-project.io/v1alpha1',
    kind: 'TokenRequest',
    metadata: { name: 'commerce', namespace: targetNamespace }
  }

  const path = `/apis/applicationconnector.kyma-project.io/v1alpha1/namespaces/${targetNamespace}/tokenrequests`

  await k8sDynamicApi.delete(tokenRequest).catch(() => { })// Ignore delete error
  await k8sDynamicApi.create(tokenRequest)
  const tokenObj = await waitForK8sObject(path, {}, (type, apiObj, watchObj) => {
    return (watchObj.object.status && watchObj.object.status.state == 'OK' && watchObj.object.status.url)
  }, 5 * 1000, "Wait for TokenRequest timeout")

  await axios.post(
    `https://${mockHost}/connection`,
    {
      token: tokenObj.status.url,
      baseUrl: `https://${mockHost}`,
      insecure: true,
    },
    {
      headers: {
        "content-type": "application/json",
      },
    }
  ).catch(expectNoAxiosErr);
}
function serviceInstanceObj(name, serviceClassExternalName) {
  return {
    apiVersion: 'servicecatalog.k8s.io/v1beta1',
    kind: 'ServiceInstance',
    metadata: { name },
    spec: { serviceClassExternalName }
  }
}
async function patchAppGatewayDeployment() {
  const commerceApplicationGatewayDeployment = await retryPromise(
    async () => {
      return k8sAppsApi.readNamespacedDeployment("commerce-application-gateway", "kyma-integration");
    },
    3,
    5000
  ).catch(expectNoK8sErr);
  expect(
    commerceApplicationGatewayDeployment.body.spec.template.spec.containers[0].args[6]
  ).to.match(/^--skipVerify/);
  const patch = [{"op": "replace", "path": "/spec/template/spec/containers/0/args/6", "value": "--skipVerify=true"}]
  const options = { "headers": { "Content-type": k8s.PatchUtils.PATCH_FORMAT_JSON_PATCH}};
  await k8sDynamicApi.requestPromise({ url: k8sDynamicApi.basePath + commerceApplicationGatewayDeployment.body.metadata.selfLink, 
    method: 'PATCH', body:patch, json:true, headers:options.headers }).catch(expectNoK8sErr);

  const patchedDeployment = await k8sAppsApi.readNamespacedDeployment("commerce-application-gateway", "kyma-integration");
  expect(
    patchedDeployment.body.spec.template.spec.containers[0].args[6]).to.equal("--skipVerify=true");
  return patchedDeployment;
}

async function ensureCommerceMockTestFixture(mockNamespace, targetNamespace) {
  const serviceBinding = {
    apiVersion: 'servicecatalog.k8s.io/v1beta1',
    kind: 'ServiceBinding',
    metadata: { name: 'commerce-binding' },
    spec: {
      instanceRef: { name: 'commerce-webservices' }
    }
  }
  const sbu = {
    apiVersion: 'servicecatalog.kyma-project.io/v1alpha1',
    kind: 'ServiceBindingUsage',
    metadata: { name: 'commerce-lastorder-sbu' },
    spec: {
      serviceBindingRef: { name: 'commerce-binding' },
      usedBy: { kind: 'serverless-function', name: 'lastorder' }
    }
  }
  await k8sApply([namespaceObj(mockNamespace), namespaceObj(targetNamespace)]);
  await k8sApply(commerceObjs);
  await k8sApply(lastorderObjs, targetNamespace, true);
  const vs = await waitForVirtualService('mocks', 'commerce-mock')
  const mockHost = vs.spec.hosts[0]
  await patchAppGatewayDeployment();
  await retryPromise(
    () => axios.get(`https://${mockHost}/local/apis`).catch(expectNoAxiosErr), 30, 3000);

  await retryPromise(() => connectMock(mockHost, targetNamespace), 3, 1000);
  await retryPromise(() => registerAllApis(mockHost), 3, 1000);

  const webServicesSC = await waitForServiceClass("webservices", targetNamespace);
  const eventsSC = await waitForServiceClass("events", targetNamespace);
  const webServicesSCExternalName = webServicesSC.spec.externalName;
  const eventsSCExternalName = eventsSC.spec.externalName;
  const serviceCatalogObjs = [
    serviceInstanceObj('commerce-webservices', webServicesSCExternalName),
    serviceInstanceObj('commerce-events', eventsSCExternalName)
  ];

  await retryPromise(() => k8sApply(serviceCatalogObjs, targetNamespace, false), 5, 2000);
  await waitForServiceInstance('commerce-webservices', targetNamespace);
  await waitForServiceInstance('commerce-events', targetNamespace);
  await k8sApply([serviceBinding], targetNamespace, false);
  await waitForServiceBinding('commerce-binding', targetNamespace);

  await k8sApply([sbu], targetNamespace);
  await waitForServiceBindingUsage('commerce-lastorder-sbu', targetNamespace)

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
  ]

}

function cleanMockTestFixture(mockNamespace, targetNamespace) {
  for (let path of getResourcePaths(mockNamespace).concat(getResourcePaths(targetNamespace))) {
    deleteAllK8sResources(path)
  }
  k8sDynamicApi.delete({
    apiVersion: 'applicationconnector.kyma-project.io/v1alpha1',
    kind: 'Application',
    metadata: {
      name: 'commerce'
    }
  })
  return deleteNamespaces([mockNamespace, targetNamespace]);

}
module.exports = {
  ensureCommerceMockTestFixture,
  sendEventAndCheckResponse,
  checkAppGatewayResponse,
  cleanMockTestFixture
};
