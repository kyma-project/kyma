const k8s = require("@kubernetes/client-node");
const {
  commerceMockYaml,
  genericServiceClass,
  serviceCatalogResources,
  mocksNamespaceYaml,
} = require("./fixtures/commerce-mock");
const { expect, config } = require("chai");
config.truncateThreshold = 0; // more verbose errors

const {
  retryPromise,
  expectNoK8sErr,
  expectNoAxiosErr,
  sleep,
  waitForTokenRequestReady,
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
const mocksNamespaceObj = k8s.loadYaml(mocksNamespaceYaml);
let mockHost;
let host;

describe("Commerce Mock tests", function () {
  this.timeout(10 * 60 * 1000);
  this.slow(1000);

  after(async function () {
    // return;
    console.log("Deleting test resources...");
    await Promise.all(
      [
        mocksNamespaceObj,
        ...commerceObjs,
        ...k8s.loadAllYaml(serviceCatalogResources("", "")),
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

  it("mocks namespace should be created or fail with reason AlreadyExists ", async function () {
    await k8sDynamicApi.create(mocksNamespaceObj).catch((e) => {
      expect(e.body.reason).to.be.equal('AlreadyExists')
    });
  })

  it("commerce mock and application should be created", async function () {
    await Promise.all(
      commerceObjs.map((obj) => k8sDynamicApi.create(obj))
    ).catch((e) => { expect(e.body.reason).to.be.equal('AlreadyExists') });
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
  });

  it("commerce mock local apis should be available", async function () {
    await retryPromise(
      () => axios.get(`https://${mockHost}/local/apis`).catch(expectNoAxiosErr), 30, 3000);
  })

  it("commerce mock should connect to Kyma", async function () {
    this.retries(4)
    const tokenRequest = {
      apiVersion: 'applicationconnector.kyma-project.io/v1alpha1',
      kind: 'TokenRequest',
      metadata: { name: 'commerce', namespace: 'default' }
    }
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
    );
  });

  it("commerce mock should register Commerce Webservices API and Events", async function () {
    this.retries(10);
    const remoteApis = await registerAllApis(mockHost, "default", watch, 5000);

    const webServicesSCExternalName = remoteApis.data.find((elem) =>
      elem.name.includes("Commerce Webservices")
    ).externalName;

    const eventsSCExternalName = remoteApis.data.find((elem) =>
      elem.name.includes("Events")
    ).externalName;

    const serviceCatalogObjs = await k8s.loadAllYaml(
      serviceCatalogResources(webServicesSCExternalName, eventsSCExternalName)
    );

    await Promise.all(
      serviceCatalogObjs.map((obj) => k8sDynamicApi.create(obj))
    ).catch((e) => {
      expect(e.body.reason).to.be.equal('AlreadyExists')
    })
  });

  it("order.created.v1 event should trigger the lastorder function", async function () {
    await sendEventAndCheckResponse(mockHost, host);
  });
});

async function sendEventAndCheckResponse(mockHost, host) {
  await retryPromise(
    async () => {
      console.log("Sending order.created event");
      await axios.post(
        `https://${mockHost}/events`,
        {
          "event-type": "order.created",
          "event-type-version": "v1",
          "event-time": "2020-09-28T14:47:16.491Z",
          data: { orderCode: "123" },
          "event-tracing": true,
        },
        {
          headers: {
            "content-type": "application/json",
          },
        }
      ).catch(e => console.dir({ status: e.response.status, data: e.response.data }));

      await sleep(500);

      console.log("Checking if event reached lambda");
      return axios.get(`https://lastorder.${host}`).then((res) => {
        expect(res.data).to.have.nested.property(
          "totalPriceWithTax.value",
          100
        );

        expect(res.data.totalPriceWithTax.value).to.equal(
          100,
          "totalPriceWithTax.value send by function is different than expected"
        );
        return res;
      });
    },
    30,
    10 * 1000
  ).catch(expectNoAxiosErr);
}

function waitForK8sObject(watch, path, query, checkFn, timeout, timeoutMsg) {
  let res
  let timer
  const result = new Promise((resolve, reject) => {
    watch.watch(path, query, (type, apiObj, watchObj) => {
      if (checkFn(type, apiObj, watchObj)) {
        if (res) {
          res.abort();
        }
        clearTimeout(timer)
        resolve(watchObj.object)
      }
    }, ()=>{}).then((r) => { res = r; timer = setTimeout(() => { res.abort(); reject(new Error(timeoutMsg)) }, timeout); })
  });
  return result;
}

async function registerAllApis(mockHost, namespace, watch, timeout = 60 * 1000) {
  const localApis = await axios.get(`https://${mockHost}/local/apis`).catch(expectNoAxiosErr);
  for (api of localApis.data) {
    await axios.post(`https://${mockHost}/local/apis/${api.id}/register`, {},
      {
        headers: {
          "content-type": "application/json",
          origin: `https://${mockHost}`,
        },
      }
    ).catch(expectNoAxiosErr)
  }
  const remoteApis = await axios.get(`https://${mockHost}/remote/apis`).catch(expectNoAxiosErr);
  expect(remoteApis.data).to.have.lengthOf(localApis.data.length)

  const path = `/apis/servicecatalog.k8s.io/v1beta1/namespaces/${namespace}/serviceclasses`
  await waitForK8sObject(watch, path, {}, (type, apiObj, watchObj) => {
    let api = remoteApis.data.find(item => item.id == watchObj.object.spec.externalID);
    if (api) {
      api.externalName = watchObj.object.spec.externalName
    }
    return remoteApis.data.every(item => item.externalName)
  }, timeout, "Wait for ServiceClasses Timeout");
  return remoteApis;
}
