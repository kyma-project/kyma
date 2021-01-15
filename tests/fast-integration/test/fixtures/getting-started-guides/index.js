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

  await createOrder(serviceDomain, order);

  await getOrders(serviceDomain, (resp) => {
    expect(resp).to.be.an("Array").of.length(1);
    expect(resp[0]).to.deep.eq(order);
  });

  await deleteAllK8sResources('/api/v1/namespaces/orders-service/pods', { labelSelector: `app=${orderService}` });

  await getOrders(
    serviceDomain,
    (resp) => {
      expect(resp).to.be.an("Array").of.length(1);
      expect(resp[0]).to.deep.eq(order);
    },
    30, // longer, because the pod has just been killed and it needs to start again
    2000
  );

  // https://kyma-project.io/docs/root/getting-started/#getting-started-connect-an-external-application
  // This is covered by commerce-mock.js test

}


async function getOrders(
  serviceDomain,
  expectFn,
  retriesLeft = 20,
  interval = 2000
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
  await waitForServiceBindingUsage(orderService, orderService);
  await deleteAllK8sResources('/api/v1/namespaces/orders-service/pods', { labelSelector: `app=${orderService}` });

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