const k8s = require("@kubernetes/client-node");
const {
  ordersServiceNamespaceYaml,
  ordersServiceMicroserviceYaml,
  addonServiceBindingServiceInstanceYaml,
  sbuYaml,
  xfMocksYaml,
} = require("./fixtures/getting-started-guides");
const { expect, config } = require("chai");
config.truncateThreshold = 0; // more verbose errors

const {
  retryPromise,
  expectNoAxiosErr,
  expectNoK8sErr,
  removeServiceInstanceFinalizer,
  removeServiceBindingFinalizer,
  sleep,
  promiseAllSettled,
  waitForTokenRequestReady,
  waitForK8sObject,
  waitForVirtualService,
  waitForServiceBinding,
  waitForServiceInstance,
} = require("../utils");

const https = require("https");
const axios = require("axios").default;

const httpsAgent = new https.Agent({
  rejectUnauthorized: false, // curl -k
});
axios.defaults.httpsAgent = httpsAgent;

const kc = new k8s.KubeConfig();
kc.loadFromDefault();

const watch = new k8s.Watch(kc);

const k8sDynamicApi = kc.makeApiClient(k8s.KubernetesObjectApi);
const k8sCRDApi = kc.makeApiClient(k8s.CustomObjectsApi);
const k8sCoreV1Api = kc.makeApiClient(k8s.CoreV1Api);

const ordersServiceNamespaceObj = k8s.loadYaml(ordersServiceNamespaceYaml);
const ordersServiceMicroserviceObj = k8s.loadAllYaml(
  ordersServiceMicroserviceYaml
);
const addonServiceBindingServiceInstanceObjs = k8s.loadAllYaml(
  addonServiceBindingServiceInstanceYaml
);
const sbuObj = k8s.loadYaml(sbuYaml);
const xfMockObjs = k8s.loadAllYaml(xfMocksYaml);

const orderService = "orders-service";

function extractFailMessages(arg) {
  return (arg || [])
    .filter((elem) => elem.status !== "fulfilled")
    .map(
      (elem) => !!elem.reason && !!elem.reason.body && elem.reason.body.message
    );
}

describe("Getting Started Guides tests", function () {
  this.timeout(10 * 60 * 1000);
  let serviceDomain;
  let host;

  const order = {
    orderCode: "762727210",
    consignmentCode: "76272725",
    consignmentStatus: "PICKUP_COMPLETE",
  };

  after(async function () {
    this.timeout(10 * 10000);

    // remove finalizers so that they don't block namespace deletion
    await removeServiceInstanceFinalizer(
      k8sCRDApi,
      "redis-service",
      orderService
    ).catch((err) => {
      // best effort finalizer deletion, no error thrown
      console.error(err.body);
    });
    await removeServiceBindingFinalizer(
      k8sCRDApi,
      orderService,
      orderService
    ).catch((err) => {
      // best effort finalizer deletion, no error thrown
      console.error(err.body);
    });

    const deletionStatuses = await promiseAllSettled(
      [
        sbuObj,
        ...addonServiceBindingServiceInstanceObjs,
        ordersServiceNamespaceObj,
        ...ordersServiceMicroserviceObj
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
    ).catch(expectNoK8sErr);

    expect(extractFailMessages(deletionStatuses)).to.have.lengthOf(0);
  });

  it("order-service namespace should be created", async function () {
    // https://kyma-project.io/docs/root/getting-started/#getting-started-create-a-namespace
    await k8sDynamicApi.create(ordersServiceNamespaceObj).catch(expectNoK8sErr);
  });

  it("order-sservice deployment, service and apirule should be created", async function () {
    // https://kyma-project.io/docs/root/getting-started/#getting-started-deploy-a-microservice-create-the-deployment
    // https://kyma-project.io/docs/root/getting-started/#getting-started-deploy-a-microservice-create-the-service
    // https://kyma-project.io/docs/root/getting-started/#getting-started-expose-the-microservice-expose-the-service
    await Promise.all(
      ordersServiceMicroserviceObj.map((obj) => k8sDynamicApi.create(obj))
    ).catch(expectNoK8sErr);
  });

  it("Addon, service instance and binding should be created", async function () {
    // https://kyma-project.io/docs/root/getting-started/#getting-started-add-the-redis-service
    // https://kyma-project.io/docs/root/getting-started/#getting-started-create-a-service-instance-for-the-redis-service
    // https://kyma-project.io/docs/root/getting-started/#getting-started-bind-the-redis-service-instance-to-the-microservice
    await Promise.all(
      addonServiceBindingServiceInstanceObjs.map((obj) =>
        k8sDynamicApi.create(obj)
      )
    ).catch(expectNoK8sErr);
    await waitForServiceInstance(watch,'redis-service',orderService);
  });

  it("APIRule should be ready", async function () {
    const apiRulePath = `/apis/gateway.kyma-project.io/v1alpha1/namespaces/${orderService}/apirules`
    waitForK8sObject(watch, apiRulePath, {}, (_type, _apiObj, watchObj) => {
      return (watchObj.object.metadata.name == orderService && watchObj.object.status.APIRuleStatus.code == "OK")
    }, 10 * 1000, "Waiting for APIRule to be ready timeout")
  });

  it("VirtualService should be created", async function () {
    // https://kyma-project.io/docs/root/getting-started/#getting-started-expose-the-microservice-call-and-test-the-microservice
    const virtualService = await waitForVirtualService(watch, orderService, orderService);
    serviceDomain = await virtualService.spec.hosts[0];
    host = serviceDomain.split(".").slice(1).join(".");
  });

  it("In-memory order-service should have volatile persistence", async function () {
    await getOrders(serviceDomain, (resp) =>
      expect(resp).to.be.an("Array").of.length(0)
    );

    await createOrder(serviceDomain, order);

    await getOrders(serviceDomain, (resp) => {
      expect(resp).to.be.an("Array").of.length(1);
      expect(resp[0]).to.deep.eq(order);
    });

    await deletePodsByLabel(orderService, `app=${orderService}`);

    await getOrders(serviceDomain, (resp) =>
      expect(resp).to.be.an("Array").of.length(0)
    );

  });

  it("ServiceBindingUsage should be created", async function () {
    // https://kyma-project.io/docs/root/getting-started/#getting-started-bind-the-redis-service-instance-to-the-microservice
    await k8sDynamicApi.create(sbuObj).catch(expectNoK8sErr);
    await waitForServiceBinding(watch, orderService, orderService);
    const secret = await retryPromise(
      async () => {
        const sec = await k8sCoreV1Api.readNamespacedSecret(
          orderService,
          orderService
        );
        return sec;
      },
      10,
      5000
    ).catch(expectNoK8sErr);

    expect(secret.body.data).to.have.property("HOST");
    expect(secret.body.data).to.have.property("PORT");
    expect(secret.body.data).to.have.property("REDIS_PASSWORD");
  });

  it("order-service should use redis persistence", async function () {

    // TODO @aerfio I think that this step with checking if secret exists is kinda redundant, would get rid of it, we already check if
    await deletePodsByLabel(orderService, `app=${orderService}`);
    // TODO jesus christ another nasty datarace, old pods that are already being deleted are still receiving traffic
    // we might have to introduce some function that not only deletes the pods, but also waits till they're completly gone
    await sleep(20000);

    await getOrders(serviceDomain, (resp) =>
      expect(resp).to.be.an("Array").of.length(0)
    );

    await createOrder(serviceDomain, order);

    await getOrders(serviceDomain, (resp) => {
      expect(resp).to.be.an("Array").of.length(1);
      expect(resp[0]).to.deep.eq(order);
    });

    await deletePodsByLabel(orderService, `app=${orderService}`);

    await getOrders(
      serviceDomain,
      (resp) => {
        expect(resp).to.be.an("Array").of.length(1);
        expect(resp[0]).to.deep.eq(order);
      },
      20, // longer, because the pod has just been killed and it needs to start again
      5000
    );

    // https://kyma-project.io/docs/root/getting-started/#getting-started-connect-an-external-application
    // This is covered by commerce-mock.js test

  });
});


async function getOrders(
  serviceDomain,
  expectFn,
  retriesLeft = 10,
  interval = 5000
) {
  return await retryPromise(
    async () => {
      return axios.get(`https://${serviceDomain}/orders`).then((res) => {
        expectFn(res.data);
        return res;
      });
    },
    retriesLeft,
    interval
  ).catch(expectNoAxiosErr);
}

async function createOrder(serviceDomain, order) {
  return await retryPromise(
    async () => {
      return axios.post(`https://${serviceDomain}/orders`, order, {
        headers: {
          "Cache-Control": "no-cache",
        },
      });
    },
    10,
    3000
  ).catch(expectNoAxiosErr);
}

async function deletePodsByLabel(namespace, label) {
  await k8sCoreV1Api
    .deleteCollectionNamespacedPod(
      namespace,
      undefined,
      undefined,
      undefined,
      undefined,
      undefined,
      label
    )
    .catch(expectNoK8sErr);
}
