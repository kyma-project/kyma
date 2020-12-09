const k8s = require("@kubernetes/client-node");
const {
  commerceMockYaml,
  appConnectorYaml,
  tokenRequestYaml,
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
const appConnectorObjs = k8s.loadAllYaml(appConnectorYaml);
const tokenRequestObj = k8s.loadYaml(tokenRequestYaml);

describe("dummy test", function () {
  this.timeout(180 * 1000); // 50s

  after(async function () {
    this.timeout(10 * 10000);
    try {
      await Promise.all(
        [tokenRequestObj, ...appConnectorObjs, ...commerceObjs].map((obj) =>
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

  it("commerce mock create", async function () {
    try {
      // we can extract namespace creation into seperate step, and _not_ fail test wgeb
      await Promise.all(commerceObjs.map((obj) => k8sDynamicApi.create(obj)));
    } catch (err) {
      expect(err.body.message).to.be.empty;
    }

    await sleep(5000); // TODO: add waiting for virtualservice

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

    // try {
    await Promise.all(appConnectorObjs.map((obj) => k8sDynamicApi.create(obj)));
    // } catch (err) {
    // expect(err.body.message).to.be.empty;
    // }

    await sleep(40000); // TODO: add retries for axios.get instead of sleep

    let response;
    try {
      response = await axios.get(`https://${mockHost}/local/apis`);
    } catch (error) {
      expect(error).to.be.empty;
    }

    expect(response.data).to.have.lengthOf(2);
    expect(response.data[0].provider).not.to.be.empty;
    const provider = response.data[0].provider;

    await sleep(5000); // TODO: introduce proper mechanism that waits till commerce-application-gateway exists

    let deploy;
    try {
      deploy = await k8sAppsApi.readNamespacedDeployment(
        "commerce-application-gateway",
        "kyma-integration"
      );
    } catch (err) {
      expect(err.body.message).to.be.empty;
    }

    expect(deploy.body.spec.template.spec.containers[0].args[6]).to.equal(
      "--skipVerify=false"
    );
    deploy.body.spec.template.spec.containers[0].args[6] = "--skipVerify=true";

    try {
      await k8sDynamicApi.patch(deploy.body);
    } catch (err) {
      expect(err.body.message).to.be.empty;
    }

    try {
      await k8sDynamicApi.create(tokenRequestObj);
    } catch (err) {
      expect(err.body.message).to.be.empty;
    }

    await sleep(10 * 1000);

    let tokenObj;
    try {
      tokenObj = await k8sDynamicApi.read(tokenRequestObj);
    } catch (err) {
      expect(err.body.message).to.be.empty;
    }

    // TODO make sure that .status.token exists first
    expect(tokenObj.body.status.token).not.to.be.empty;

    const token = tokenObj.body.status.token;

    await sleep(15 * 1000);

    const host = "local.kyma.dev"; // TODO: extract this from some secret/configmap/ parse some virtualservice
    try {
      await axios.post(
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
    } catch (error) {
      // https://github.com/axios/axios#handling-errors
      expect(error.response.data).to.deep.eq({});
    }
  });
});

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}
