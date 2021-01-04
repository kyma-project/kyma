const k8s = require("@kubernetes/client-node");
const {
  commerceMockYaml,
  serviceCatalogResources,
  mocksNamespaceYaml,
  lastorderFunctionYaml,
} = require("./fixtures/commerce-mock");
const { expect, config } = require("chai");
config.truncateThreshold = 0; // more verbose errors

const {
  retryPromise,
  expectNoK8sErr,
  expectNoAxiosErr,
  sleep,
  k8sApply,
  waitForK8sObject,
  waitForServiceClass,
  waitForServiceInstance,
  waitForServiceBinding
} = require("../utils");

const https = require("https");
const axios = require("axios").default;
const httpsAgent = new https.Agent({
  rejectUnauthorized: false, // curl -k
});
axios.defaults.httpsAgent = httpsAgent;

const kc = new k8s.KubeConfig();
kc.loadFromDefault();

const k8sDynamicApi = kc.makeApiClient(k8s.KubernetesObjectApi);
const k8sAppsApi = kc.makeApiClient(k8s.AppsV1Api);
const watch = new k8s.Watch(kc);

const commerceObjs = k8s.loadAllYaml(commerceMockYaml);
const lastorderObjs = k8s.loadAllYaml(lastorderFunctionYaml);
const mocksNamespaceObj = k8s.loadYaml(mocksNamespaceYaml);

const tokenRequest = {
  apiVersion: 'applicationconnector.kyma-project.io/v1alpha1',
  kind: 'TokenRequest',
  metadata: { name: 'commerce', namespace: 'default' }
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

describe("Commerce Mock tests", function () {
  this.timeout(10 * 60 * 1000);
  this.slow(5000);
  let mockHost;
  let host;

  after(async function () {
    // return;
    await Promise.all(
      [
        mocksNamespaceObj,
        ...commerceObjs,
        ...lastorderObjs,
        ...k8s.loadAllYaml(serviceCatalogResources("", "")),
        sbu,
        tokenRequest
      ].map((obj) =>
        k8sDynamicApi.delete(
          obj,
          "true",
          undefined,
          undefined,
          undefined,
          "Foreground" // namespaces seem to ignore this
        )
      )
    ).catch((e) => {
      expect(e.body.reason).to.be.equal('NotFound')
    });
  });

  it("mocks namespace should be created", async function () {
    k8sApply(k8sDynamicApi, [mocksNamespaceObj]);
  })

  it("commerce mock and application should be created", async function () {
    k8sApply(k8sDynamicApi, commerceObjs);
  });

  it("lastorder function should be created", async function () {
    k8sApply(k8sDynamicApi, lastorderObjs);
  });

  it("commerce mock should be exposed with VirtualService", async function () {
    const path = `/apis/networking.istio.io/v1beta1/namespaces/mocks/virtualservices`
    const query = { labelSelector: "apirule.gateway.kyma-project.io/v1alpha1=commerce-mock.mocks" }
    const vs = await waitForK8sObject(watch, path, query, (type, apiObj, watchObj) => {
      return watchObj.object.spec.hosts && watchObj.object.spec.hosts.length == 1
    }, 30 * 1000, "Wait for VirtualService Timeout");
    mockHost = vs.spec.hosts[0]
    host = mockHost.split(".").slice(1).join(".");
  });

  it("commerce-application gateway should be deployed", async function () {
    this.retries(3);
    const commerceApplicationGatewayDeployment = await retryPromise(
      async () => {
        return k8sAppsApi.readNamespacedDeployment("commerce-application-gateway", "kyma-integration");
      },
      10,
      5000
    ).catch(expectNoK8sErr);

    expect(
      commerceApplicationGatewayDeployment.body.spec.template.spec.containers[0].args[6]
    ).to.match(/^--skipVerify/);
    commerceApplicationGatewayDeployment.body.spec.template.spec.containers[0].args[6] =
      "--skipVerify=true";

    await k8sDynamicApi
      .patch(commerceApplicationGatewayDeployment.body)
      .catch(expectNoK8sErr);

    const patchedDeployment = await k8sAppsApi.readNamespacedDeployment("commerce-application-gateway", "kyma-integration");
    expect(
      patchedDeployment.body.spec.template.spec.containers[0].args[6]).to.equal("--skipVerify=true");
  });

  it("commerce mock local apis should be available", async function () {
    await retryPromise(
      () => axios.get(`https://${mockHost}/local/apis`).catch(expectNoAxiosErr), 30, 3000);
  })

  it("commerce mock should connect to Kyma", async function () {
    this.retries(4)
    const path = `/apis/applicationconnector.kyma-project.io/v1alpha1/namespaces/default/tokenrequests`

    await k8sDynamicApi.delete(tokenRequest).catch(() => { })// Ignore delete error
    await k8sDynamicApi.create(tokenRequest)
    const tokenObj = await waitForK8sObject(watch, path, {}, (type, apiObj, watchObj) => {
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
  });

  it("Commerce mock should register Websevices API and Events API", async function () {
    this.retries(3);
    const remoteApis = await registerAllApis(mockHost, "default", watch, 30 * 1000);

  })

  it("Commerce Webservices API and Events service instances should be ready", async function () {
    const webServicesSC = await waitForServiceClass(watch, "webservices");
    const eventsSC = await waitForServiceClass(watch, "events");

    const webServicesSCExternalName = webServicesSC.spec.externalName;
    const eventsSCExternalName = eventsSC.spec.externalName;
    await sleep(5000); // Some issue with creating service instances when service class doesn't have service plans yet
    const serviceCatalogObjs = k8s.loadAllYaml(
      serviceCatalogResources(webServicesSCExternalName, eventsSCExternalName)
    );
    await retryPromise(() => k8sApply(k8sDynamicApi, serviceCatalogObjs, false), 5, 2000);

    await waitForServiceInstance(watch, 'commerce-webservices');
    await waitForServiceInstance(watch, 'commerce-events');
    await waitForServiceBinding(watch, 'commerce-binding');

  });


  it("service binding usage for function should be ready", async function () {
    k8sDynamicApi.create(sbu);
    const path = '/apis/servicecatalog.kyma-project.io/v1alpha1/namespaces/default/servicebindingusages';
    await waitForK8sObject(watch, path, {}, (type, apiObj, watchObj) => {
      if (watchObj.object.metadata.name == 'commerce-lastorder-sbu' && watchObj.object.status.conditions) {
        return watchObj.object.status.conditions.some((c) => (c.type == 'Ready' && c.status == 'True'))
      }
      return false;
    }, 90 * 1000, "Waiting for ServiceBindingUsage to be ready timeout");
  });

  it("function should reach Commerce mock API through app gateway", async function () {
    let res = await retryPromise(() => axios.post(`https://lastorder.${host}`, { orderCode: "123" }), 10, 2000);
    expect(res.data).to.have.nested.property("order.totalPriceWithTax.value", 100);
  })

  it("order.created.v1 event should trigger the lastorder function", async function () {
    await sendEventAndCheckResponse(mockHost, host);
  });
});


async function sendEventAndCheckResponse(mockHost, host) {
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

async function registerAllApis(mockHost, namespace, watch, timeout = 60 * 1000) {
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
