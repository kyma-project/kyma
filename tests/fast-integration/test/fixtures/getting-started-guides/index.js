const k8s = require("@kubernetes/client-node");
const fs = require("fs");
const path = require("path");

const ordersServiceNamespaceYaml = fs.readFileSync(
  path.join(__dirname, "./ns.yaml"),
  {
    encoding: "utf8",
  }
);

const ordersServiceMicroserviceYaml = fs.readFileSync(
  path.join(__dirname, "./microservice.yaml"),
  {
    encoding: "utf8",
  }
);

const addonServiceBindingServiceInstanceYaml = fs.readFileSync(
  path.join(__dirname, "./redis-addon-sb-si.yaml"),
  {
    encoding: "utf8",
  }
);

const sbuYaml = fs.readFileSync(path.join(__dirname, "./redis-sbu.yaml"), {
  encoding: "utf8",
});

const { expect, config } = require("chai");
config.truncateThreshold = 0; // more verbose errors

const {
  retryPromise,
  expectNoAxiosErr,
  waitForK8sObject,
  waitForVirtualService,
  waitForServiceBinding,
  waitForServiceInstance,
  k8sApply,
  deleteAllK8sResources,
  waitForServiceBindingUsage,
  deleteNamespaces,
} = require("../../../utils");

const https = require("https");
const axios = require("axios").default;

const httpsAgent = new https.Agent({
  rejectUnauthorized: false, // curl -k
});
axios.defaults.httpsAgent = httpsAgent;


const ordersServiceNamespaceObj = k8s.loadYaml(ordersServiceNamespaceYaml);
const ordersServiceMicroserviceObj = k8s.loadAllYaml(
  ordersServiceMicroserviceYaml
);
const addonServiceBindingServiceInstanceObjs = k8s.loadAllYaml(
  addonServiceBindingServiceInstanceYaml
);
const sbuObj = k8s.loadYaml(sbuYaml);

const orderService = "orders-service";

const order = {
  orderCode: "762727210",
  consignmentCode: "76272725",
  consignmentStatus: "PICKUP_COMPLETE",
};


async function verifyOrderPersisted() {
  const virtualService = await waitForVirtualService(orderService, orderService);
  const serviceDomain = await virtualService.spec.hosts[0];

  await retryPromise(async () => {
    await createOrder(serviceDomain, order);
    return findOrder(serviceDomain, order);
  }, 30, 2000).catch(e => {throw new Error("Error during creating order: "+e)});

  await deleteAllK8sResources('/api/v1/namespaces/orders-service/pods', { labelSelector: `app=${orderService}` });

  await retryPromise(() => findOrder(serviceDomain, order), 30, 2000);

  // https://kyma-project.io/docs/root/getting-started/#getting-started-connect-an-external-application
  // This is covered by commerce-mock.js test
}


async function findOrder(serviceDomain, order) {
  const result = await axios.get(`https://${serviceDomain}/orders`)
  if (result.data && result.data.length) {
    const createdOrder = result.data.find((o) => o.orderCode == order.orderCode);
    if (createdOrder) {
      return createdOrder;
    }
  }
  throw new Error("Order not found: "+order.orderCode)
}

async function createOrder(serviceDomain, order) {
  return  axios.post(`https://${serviceDomain}/orders`, order, {
    headers: {
      "Cache-Control": "no-cache",
    },
  }).catch(err => { if (err.response.status != 409) throw new Error("Cannot create the order. Error: " + err) });
}

function waitForPodWithSbuToBeReady(sbu) {
  return waitForK8sObject('/api/v1/namespaces/orders-service/pods', { labelSelector: `app=${orderService}` }
    , (_type, _apiObj, watchObj) => {
      return Object.keys(watchObj.object.metadata.labels).includes('use-' + sbu.metadata.uid) &&
        watchObj.object.metadata.name.startsWith(orderService) && watchObj.object.status.conditions
        && watchObj.object.status.conditions.some((c) => (c.type == 'Ready' && c.status == 'True'))
    }
    , 60 * 1000, "Waiting for pods with injected redis service timeout");
}

async function ensureGettingStartedTestFixture() {
  await k8sApply([ordersServiceNamespaceObj]);
  await k8sApply(ordersServiceMicroserviceObj, orderService);
  await k8sApply(addonServiceBindingServiceInstanceObjs, orderService);
  const apiRulePath = `/apis/gateway.kyma-project.io/v1alpha1/namespaces/${orderService}/apirules`
  await waitForK8sObject(apiRulePath, {}, (_type, _apiObj, watchObj) => {
    return (watchObj.object.metadata.name == orderService && watchObj.object.status.APIRuleStatus.code == "OK")
  }, 10 * 1000, "Waiting for APIRule to be ready timeout")
  await waitForVirtualService(orderService, orderService);
  await waitForServiceInstance('redis-service', orderService);
  await waitForServiceBinding(orderService, orderService);
  await k8sApply([sbuObj], orderService);
  const sbu = await waitForServiceBindingUsage(orderService, orderService);
  await waitForPodWithSbuToBeReady(sbu);
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
  ]
}

function cleanGettingStartedTestFixture(wait = true) {
  for (let path of getResourcePaths(orderService)) {
    deleteAllK8sResources(path)
  }
  return deleteNamespaces([orderService], wait);
}

module.exports = {
  ensureGettingStartedTestFixture,
  verifyOrderPersisted,
  cleanGettingStartedTestFixture
}