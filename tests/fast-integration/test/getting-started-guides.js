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

    console.log("Deleting test resources...");
    const deletionStatuses = await promiseAllSettled(
      [
        sbuObj,
        ...addonServiceBindingServiceInstanceObjs,
        ordersServiceNamespaceObj,
        ...ordersServiceMicroserviceObj,
        ...xfMockObjs,
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

  // TODO: check if we can split this one big "it" into several smaller ones
  // mocha should preserve the order of test inside "describe", but I'm not sure
  it("should pass with 😄", async function () {
    // https://kyma-project.io/docs/root/getting-started/#getting-started-create-a-namespace
    console.log("Creating orders-service namespace...");
    await k8sDynamicApi.create(ordersServiceNamespaceObj).catch(expectNoK8sErr);

    // https://kyma-project.io/docs/root/getting-started/#getting-started-deploy-a-microservice-create-the-deployment
    // https://kyma-project.io/docs/root/getting-started/#getting-started-deploy-a-microservice-create-the-service
    // https://kyma-project.io/docs/root/getting-started/#getting-started-expose-the-microservice-expose-the-service
    console.log("Creating orders-service deployment, service and apirule...");
    await Promise.all(
      ordersServiceMicroserviceObj.map((obj) => k8sDynamicApi.create(obj))
    ).catch(expectNoK8sErr);

    // https://kyma-project.io/docs/root/getting-started/#getting-started-add-the-redis-service
    // https://kyma-project.io/docs/root/getting-started/#getting-started-create-a-service-instance-for-the-redis-service
    // https://kyma-project.io/docs/root/getting-started/#getting-started-bind-the-redis-service-instance-to-the-microservice
    console.log("Creating addon, serviceinstance, servicebinding...");
    await Promise.all(
      addonServiceBindingServiceInstanceObjs.map((obj) =>
        k8sDynamicApi.create(obj)
      )
    ).catch(expectNoK8sErr);

    console.log("Waiting for apirule to be ready...");
    await waitForApiRuleReady();

    // https://kyma-project.io/docs/root/getting-started/#getting-started-expose-the-microservice-call-and-test-the-microservice
    const serviceDomain = await getServiceDomain();
    const host = serviceDomain.split(".").slice(1).join(".");
    console.log(`Host: https://${host}`);
    console.log(`Service Domain: https://${serviceDomain}`);

    await getOrders(serviceDomain, (resp) =>
      expect(resp).to.be.an("Array").of.length(0)
    );

    const order = {
      orderCode: "762727210",
      consignmentCode: "76272725",
      consignmentStatus: "PICKUP_COMPLETE",
    };
    await createOrder(serviceDomain, order);

    await getOrders(serviceDomain, (resp) => {
      expect(resp).to.be.an("Array").of.length(1);
      expect(resp[0]).to.deep.eq(order);
    });

    await deletePodsByLabel(orderService, `app=${orderService}`);

    await getOrders(serviceDomain, (resp) =>
      expect(resp).to.be.an("Array").of.length(0)
    );

    // https://kyma-project.io/docs/root/getting-started/#getting-started-bind-the-redis-service-instance-to-the-microservice
    console.log("Creating ServiceBindingUsage...");
    await k8sDynamicApi.create(sbuObj).catch(expectNoK8sErr);

    await waitForSbuReady(orderService, orderService);

    // TODO @aerfio I think that this step with checking if secret exists is kinda redundant, would get rid of it, we already check if
    const secret = await retryPromise(
      async () => {
        console.log("Waiting for SBU's secret to appear");
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
    console.log(
      "Creating addonconfiguration, serviceinstance, application and tokenrequest"
    );
    await Promise.all(xfMockObjs.map((obj) => k8sDynamicApi.create(obj))).catch(
      expectNoK8sErr
    );
  });
  // https://kyma-project.io/docs/root/getting-started/#getting-started-connect-an-external-application-connect-events
  // TODO: this step requires us to use UI, needs to be handled somehow here in nodejs
});

async function waitForApiRuleReady() {
  await retryPromise(
    async () => {
      return k8sCRDApi
        .getNamespacedCustomObject(
          "gateway.kyma-project.io",
          "v1alpha1",
          orderService,
          "apirules",
          orderService
        )
        .then((res) => {
          expect(res.body).to.have.nested.property("status.APIRuleStatus.code");
          expect(res.body.status.APIRuleStatus.code).to.equal("OK");
          return res;
        });
    },
    10,
    1000
  ).catch(expectNoK8sErr);
}

async function waitForSbuReady(
  name,
  namespace,
  retriesLeft = 40,
  interval = 5000
) {
  await retryPromise(
    async () => {
      console.log("Waiting for SBU to get ready");
      return k8sCRDApi
        .getNamespacedCustomObject(
          "servicecatalog.kyma-project.io",
          "v1alpha1",
          namespace,
          "servicebindingusages",
          name
        )
        .then((res) => {
          expect(res.body).to.have.nested.property("status.conditions");
          const condition = res.body.status.conditions.find(
            (elem) => elem.type === "Ready"
          );
          expect(condition.status).to.equal("True");
          return res;
        });
    },
    retriesLeft,
    interval
  ).catch(expectNoK8sErr);
}

async function getServiceDomain() {
  const virtualservice = await retryPromise(
    async () => {
      return k8sCRDApi
        .listNamespacedCustomObject(
          "networking.istio.io",
          "v1beta1",
          orderService,
          "virtualservices",
          "true",
          undefined,
          undefined,
          `apirule.gateway.kyma-project.io/v1alpha1=${orderService}.${orderService}`
        )
        .then((res) => {
          expect(res.body).to.have.property("items");
          expect(res.body.items).to.have.lengthOf(1);
          expect(res.body.items[0]).to.have.nested.property("spec.hosts");
          expect(res.body.items[0].spec.hosts).to.have.lengthOf(1);
          expect(res.body.items[0].spec.hosts[0]).not.to.be.empty;
          return res;
        });
    },
    10,
    1000
  ).catch(expectNoK8sErr);

  return virtualservice.body.items[0].spec.hosts[0];
}

async function getOrders(
  serviceDomain,
  expectFn,
  retriesLeft = 10,
  interval = 5000
) {
  return await retryPromise(
    async () => {
      console.log(`Calling https://${serviceDomain}/orders`);
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
      console.log(
        `Adding new order to https://${serviceDomain}/orders; order: ${JSON.stringify(
          order
        )}`
      );
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
  console.log(`Deleting pods in ${namespace} namespace by label ${label}`);
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
