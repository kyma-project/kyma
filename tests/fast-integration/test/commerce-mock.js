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
const k8sCRDApi = kc.makeApiClient(k8s.CustomObjectsApi);
const k8sAppsApi = kc.makeApiClient(k8s.AppsV1Api);

const commerceObjs = k8s.loadAllYaml(commerceMockYaml);
const mocksNamespaceObj = k8s.loadYaml(mocksNamespaceYaml);

describe("Commerce Mock tests", function () {
  this.timeout(10 * 60 * 1000);

  after(async function () {
    this.timeout(10 * 10000);

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
    ).catch(expectNoK8sErr);
  });

  // TODO: check if we can split this one big "it" into several smaller ones
  // mocha should preserve the order of test inside "describe" block, but I'm not sure
  it("should pass with 😄", async function () {
    console.log("Creating mocks namespace...");
    await k8sDynamicApi.create(mocksNamespaceObj).catch(expectNoK8sErr); // we can extract namespace creation into seperate step and ignore AlreadyExists error

    console.log("Creating commerce resources...");
    await Promise.all(
      commerceObjs.map((obj) => k8sDynamicApi.create(obj))
    ).catch(expectNoK8sErr);

    console.log("Waiting for virtual service to be ready...");

    const mockHost = await getMockVirtualHost();
    const host = mockHost.split(".").slice(1).join(".");
    console.log(`Host: https://${host}`);
    console.log(`Mock host: https://${mockHost}`);

    // @aerfio tests do not seem to pass without calling https://${mockHost}/local/apis first
    // @aerfio discussed with PB - it's probably a bug in varkes
    await retryPromise(
      async () => {
        console.log(`Calling https://${mockHost}/local/apis`);
        return axios.get(`https://${mockHost}/local/apis`).then((res) => {
          expect(res.data).to.have.lengthOf(2);
          expect(res.data[0].provider).not.to.be.empty;
          return res;
        });
      },
      40, // sometimes we need to wait this long, especially in Gardener on Azure
      5000
    ).catch(expectNoAxiosErr);

    const commerceApplicationGatewayDeployment = await retryPromise(
      async () => {
        console.log(
          "Waiting for commerce-application-gateway deployment to appear"
        );
        return k8sAppsApi.readNamespacedDeployment(
          "commerce-application-gateway",
          "kyma-integration"
        );
      },
      10,
      5000
    ).catch(expectNoK8sErr);

    expect(
      commerceApplicationGatewayDeployment.body.spec.template.spec.containers[0]
        .args[6]
    ).to.equal("--skipVerify=false");
    commerceApplicationGatewayDeployment.body.spec.template.spec.containers[0].args[6] =
      "--skipVerify=true";

    console.log("Patching commerce-application-gateway deployment");
    await k8sDynamicApi
      .patch(commerceApplicationGatewayDeployment.body)
      .catch(expectNoK8sErr);

    const tokenObj = await waitForTokenRequestReady(
      k8sCRDApi,
      "commerce",
      "default",
      15,
      5000
    );

    expect(tokenObj.body).to.have.nested.property("status.token");
    expect(tokenObj.body.status.token).not.to.be.empty;

    await retryPromise(
      async () => {
        console.log(`Connecting to https://${mockHost}/connection`);
        return axios.post(
          `https://${mockHost}/connection`,
          {
            token: tokenObj.body.status.url,
            baseUrl: `https://${mockHost}`,
            insecure: true,
          },
          {
            headers: {
              "content-type": "application/json",
            },
          }
        );
      },
      30, // azure+gardener combo ❤️
      5000
    ).catch(expectNoAxiosErr);

    await registerApis(
      `https://${mockHost}/local/apis/Commerce%20Webservices/register`,
      mockHost
    );

    const commerceWebservicesResp = await listRemoteApis(mockHost);
    expect(commerceWebservicesResp.data).to.have.length.above(0);
    const commerceWebservicesID = commerceWebservicesResp.data.find((elem) =>
      elem.name.includes("Commerce Webservices")
    ).id;
    expect(commerceWebservicesID).not.to.be.empty;

    await registerApis(
      `https://${mockHost}/local/apis/Events/register`,
      mockHost
    );

    const remoteApis = await listRemoteApis(mockHost);

    expect(remoteApis.data).to.have.length.above(1);
    const commerceEventsID = remoteApis.data.find((elem) =>
      elem.name.includes("Events")
    ).id;
    expect(commerceEventsID).not.to.be.empty;

    const webServicesServiceClass = await retryPromise(
      async () => {
        console.log("Reading Web Services service class");
        return k8sDynamicApi.read(
          k8s.loadYaml(genericServiceClass(commerceWebservicesID, "default"))
        );
      },
      20,
      5000
    ).catch(expectNoK8sErr);

    expect(webServicesServiceClass.body).to.have.nested.property(
      "spec.externalName"
    );
    const webServicesSCExternalName =
      webServicesServiceClass.body.spec.externalName;

    const eventsServiceClass = await retryPromise(
      async () => {
        console.log("Reading Events service class");
        return k8sDynamicApi.read(
          k8s.loadYaml(genericServiceClass(commerceEventsID, "default"))
        );
      },
      10,
      5000
    ).catch(expectNoK8sErr);

    expect(eventsServiceClass.body).to.have.nested.property(
      "spec.externalName"
    );
    const eventsSCExternalName = eventsServiceClass.body.spec.externalName;

    const serviceCatalogObjs = k8s.loadAllYaml(
      serviceCatalogResources(webServicesSCExternalName, eventsSCExternalName)
    );

    console.log("Creating Service Catalog resources");
    await Promise.all(
      serviceCatalogObjs.map((obj) => k8sDynamicApi.create(obj))
    ).catch(expectNoK8sErr);

    await sendEventAndCheckResponse(mockHost, host);
    console.log("Done!");
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
      );

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

async function getMockVirtualHost() {
  const virtualservice = await retryPromise(
    async () => {
      return k8sCRDApi
        .listNamespacedCustomObject(
          "networking.istio.io",
          "v1beta1",
          "mocks",
          "virtualservices",
          "true",
          undefined,
          undefined,
          "apirule.gateway.kyma-project.io/v1alpha1=commerce-mock.mocks"
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

async function listRemoteApis(mockHost, retriesLeft = 10, interval = 3000) {
  return await retryPromise(
    async () => {
      console.log("Listing remote apis");
      return axios.get(`https://${mockHost}/remote/apis`);
    },
    retriesLeft,
    interval
  ).catch(expectNoAxiosErr);
}

async function registerApis(url, mockHost, retriesLeft = 15, interval = 5000) {
  await retryPromise(
    async () => {
      console.log(`Registering Apis, calling ${url}`);
      return axios.post(
        url,
        {},
        {
          headers: {
            "content-type": "application/json",
            origin: `https://${mockHost}`,
          },
        }
      );
    },
    retriesLeft,
    interval
  ).catch(expectNoAxiosErr);
}
