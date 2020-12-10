const k8s = require("@kubernetes/client-node");
const {
  commerceMockYaml,
  tokenRequestYaml,
  genericServiceClass,
  serviceCatalogResources,
  mocksNamespaceYaml,
} = require("./fixtures");
const { expect, config } = require("chai");
config.truncateThreshold = 0;

const https = require("https");
const axios = require("axios").default;
const httpsAgent = new https.Agent({
  rejectUnauthorized: false, // curl -k
});
axios.defaults.httpsAgent = httpsAgent; // create separate axios instance with that httpsAgent https://github.com/axios/axios#custom-instance-defaults
const kc = new k8s.KubeConfig();
kc.loadFromDefault();

const k8sDynamicApi = kc.makeApiClient(k8s.KubernetesObjectApi);
const k8sCRDApi = kc.makeApiClient(k8s.CustomObjectsApi);
const k8sAppsApi = kc.makeApiClient(k8s.AppsV1Api);

const commerceObjs = k8s.loadAllYaml(commerceMockYaml);
const tokenRequestObj = k8s.loadYaml(tokenRequestYaml);
const mocksNamespaceObj = k8s.loadYaml(mocksNamespaceYaml);

function retryPromise(fn, retriesLeft = 3, interval = 200) {
  return new Promise((resolve, reject) => {
    return fn()
      .then(resolve)
      .catch((error) => {
        if (retriesLeft === 1) {
          // reject('maximum retries exceeded');
          reject(error);
          return;
        }

        setTimeout(() => {
          console.log("retriesLeft: ", retriesLeft);
          // Passing on "reject" is the important part
          retryPromise(fn, retriesLeft - 1, interval).then(resolve, reject);
        }, interval);
      });
  });
}

describe("Commerce Mock tests", function () {
  this.timeout(300 * 1000); // 50s

  after(async function () {
    this.timeout(10 * 10000);
    try {
      console.log("Deleting test resources...");
      await Promise.all(
        [
          mocksNamespaceObj,
          tokenRequestObj,
          ...commerceObjs,
          ...k8s.loadAllYaml(serviceCatalogResources("", "")), // hope it'll delete those resources, even though .spec.externalName=""
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
      );
    } catch (err) {
      console.error(err.body.message); // prints only first failed promise, might want to use Promise.allSettled
    }
  });

  it("should pass with ", async function () {
    try {
      console.log("Creating mocks namespace...");
      await k8sDynamicApi.create(mocksNamespaceObj);
      // we can extract namespace creation into seperate step and ignore AlreadyExists error
      console.log("Creating commerce resources...");
      await Promise.all(commerceObjs.map((obj) => k8sDynamicApi.create(obj)));
    } catch (err) {
      // console.error(err);
      expect(err.body.message).to.be.empty;
    }

    await sleep(5000); // TODO: add waiting for virtualservice

    console.log("Waiting for virtual service to be ready...");
    let virtualservice;
    try {
      virtualservice = await k8sCRDApi.listNamespacedCustomObject(
        "networking.istio.io",
        "v1beta1",
        "mocks",
        "virtualservices",
        "true",
        undefined,
        undefined,
        "apirule.gateway.kyma-project.io/v1alpha1=commerce-mock.mocks"
      );
    } catch (err) {
      console.error(err.body.message);
    }

    expect(virtualservice.body.items).to.have.lengthOf(1);
    expect(virtualservice.body.items[0].spec.hosts).to.have.lengthOf(1);
    expect(virtualservice.body.items[0].spec.hosts[0]).not.to.be.empty;

    const mockHost = virtualservice.body.items[0].spec.hosts[0];
    const host = mockHost.split(".").slice(1).join(".");
    console.log(`Host: ${host}`);
    console.log(`Mock host: ${mockHost}`);

    let response;
    try {
      response = await retryPromise(
        () => {
          console.log(`Calling https://${mockHost}/local/apis`);
          return axios.get(`https://${mockHost}/local/apis`).then((res) => {
            expect(res.data).to.have.lengthOf(2);
            expect(res.data[0].provider).not.to.be.empty;
            return res;
          });
        },
        10,
        5000
      );
    } catch (err) {
      expect(err).to.be.empty;
    }

    // let response;
    // try {
    //   console.log(`Calling https://${mockHost}/local/apis`);
    //   response = await axios.get(`https://${mockHost}/local/apis`);
    // } catch (error) {
    //   expect(error).to.be.empty;
    // }

    expect(response.data).to.have.lengthOf(2);
    expect(response.data[0].provider).not.to.be.empty;
    // TODO: discuss with PB whether it's needed
    // TODO2: tests do not seem to pass without it, but we do not use provider variable anywhere, needs to be discussed
    const provider = response.data[0].provider;

    await sleep(5000); // TODO: introduce proper mechanism that waits till commerce-application-gateway exists

    let commerceApplicationGatewayDeployment;
    try {
      commerceApplicationGatewayDeployment = await k8sAppsApi.readNamespacedDeployment(
        "commerce-application-gateway",
        "kyma-integration"
      );
    } catch (err) {
      expect(err.body.message).to.be.empty;
    }

    expect(
      commerceApplicationGatewayDeployment.body.spec.template.spec.containers[0]
        .args[6]
    ).to.equal("--skipVerify=false");
    commerceApplicationGatewayDeployment.body.spec.template.spec.containers[0].args[6] =
      "--skipVerify=true";

    try {
      console.log("Patching commerce-application-gateway deployment");
      await k8sDynamicApi.patch(commerceApplicationGatewayDeployment.body);
    } catch (err) {
      expect(err.body.message).to.be.empty;
    }

    try {
      console.log("Creating TokenRequest");
      await k8sDynamicApi.create(tokenRequestObj);
    } catch (err) {
      expect(err.body.message).to.be.empty;
    }

    await sleep(10 * 1000);

    let tokenObj;
    try {
      console.log("Reading TokenRequest .status.token");
      tokenObj = await k8sDynamicApi.read(tokenRequestObj);
    } catch (err) {
      expect(err.body.message).to.be.empty;
    }

    // TODO make sure that .status.token exists first
    expect(tokenObj.body.status.token).not.to.be.empty;

    const token = tokenObj.body.status.token;

    try {
      await retryPromise(
        () => {
          console.log(`Conneting to https://${mockHost}/connection`);
          return axios.post(
            `https://${mockHost}/connection`,
            {
              token: `https://connector-service.${host}/v1/applications/signingRequests/info?token=${token}`,
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
        10,
        3000
      );
    } catch (error) {
      // https://github.com/axios/axios#handling-errors
      expect(error.response.data).to.deep.eq({});
    }

    try {
      await retryPromise(
        () => {
          console.log("Registering Commerce Webservices");
          return axios.post(
            `https://${mockHost}/local/apis/Commerce%20Webservices/register`,
            {},
            {
              headers: {
                "content-type": "application/json",
                origin: `https://${mockHost}`,
              },
            }
          );
        },
        10,
        3000
      );
    } catch (err) {
      expect(err.response.data).to.deep.eq({});
    }

    // await sleep(10000);

    let commerceWebservicesResp;
    try {
      commerceWebservicesResp = await retryPromise(
        () => {
          console.log("Listing remote apis");
          return axios.get(`https://${mockHost}/remote/apis`);
        },
        10,
        3000
      );
    } catch (err) {
      expect(err.response.data).to.deep.eq({});
    }

    expect(commerceWebservicesResp.data).to.have.length.above(0);
    const commerceWebservicesID = commerceWebservicesResp.data.find((elem) =>
      elem.name.includes("Commerce Webservices")
    ).id;
    expect(commerceWebservicesID).not.to.be.empty;

    try {
      await retryPromise(
        () => {
          console.log("Registering Events");
          return axios.post(
            `https://${mockHost}/local/apis/Events/register`, // TODO: can we do this in parallel to registering commerce webservices?
            {},
            {
              headers: {
                "content-type": "application/json",
                origin: `https://${mockHost}`,
              },
            }
          );
        },
        10,
        3000
      );
    } catch (err) {
      expect(err.response.data).to.deep.eq({});
    }

    try {
      commerceWebservicesResp = await retryPromise(
        () => {
          console.log("Listing remote apis");
          return axios.get(`https://${mockHost}/remote/apis`);
        },
        10,
        3000
      );
    } catch (err) {
      expect(err.response.data).to.deep.eq({});
    }

    expect(commerceWebservicesResp.data).to.have.length.above(1);
    const commerceEventsID = commerceWebservicesResp.data.find((elem) =>
      elem.name.includes("Events")
    ).id;
    expect(commerceEventsID).not.to.be.empty;

    let webServicesServiceClass;
    try {
      webServicesServiceClass = await retryPromise(
        () => {
          console.log("Reading Web Services service class");
          return k8sDynamicApi.read(
            k8s.loadYaml(genericServiceClass(commerceWebservicesID, "default"))
          );
        },
        10,
        3000
      );
    } catch (err) {
      expect(err.body.message).to.be.empty;
    }

    const webServicesSCExternalName =
      webServicesServiceClass.body.spec.externalName;

    let eventsServiceClass;
    try {
      console.log("Reading Events service class");
      eventsServiceClass = await k8sDynamicApi.read(
        k8s.loadYaml(genericServiceClass(commerceEventsID, "default"))
      );
    } catch (err) {
      expect(err.body.message).to.be.empty;
    }

    const eventsSCExternalName = eventsServiceClass.body.spec.externalName; // TODO: check if .spec.externalName exists first

    const serviceCatalogObjs = k8s.loadAllYaml(
      serviceCatalogResources(webServicesSCExternalName, eventsSCExternalName)
    );

    try {
      console.log("Creating Service Catalog resources");
      await Promise.all(
        serviceCatalogObjs.map((obj) => k8sDynamicApi.create(obj))
      );
    } catch (err) {
      expect(err.body.message).to.be.empty;
    }

    console.time("waiting for sbu...");

    let functionResp;
    try {
      functionResp = await retryPromise(
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
            expect(res.data).to.have.property("totalPriceWithTax.value", 100);
            console.log(res.data);
            return res;
          });
        },
        30,
        10000
      );
    } catch (err) {
      expect(err.response.data).to.deep.eq({});
    }
    console.timeEnd("waiting for sbu...");
    expect(functionResp.data.totalPriceWithTax.value).to.equal(100);
    console.log("Done!");
  });
});

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}
