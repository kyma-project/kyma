const k8s = require("@kubernetes/client-node");
const { deploymentYamlString } = require("./fixtures");
const { expect } = require("chai");

const kc = new k8s.KubeConfig();
kc.loadFromDefault();

const k8sDynamicApi = kc.makeApiClient(k8s.KubernetesObjectApi);
const deployObj = k8s.loadYaml(deploymentYamlString);

describe("dummy test", function () {
  this.timeout(10000);

  before(async function () {
    try {
      await k8sDynamicApi.delete(deployObj);
    } catch (error) {
      console.warn(error.body.message);
    }
  });

  it("create deployment", async function () {
    try {
      await k8sDynamicApi.create(deployObj);
    } catch (error) {
      expect(error.body.message).to.be.empty;
    }

    deployObj.spec.replicas = 3;

    let data;
    try {
      data = await k8sDynamicApi.patch(deployObj);
    } catch (error) {
      expect(error.body.message).to.be.empty;
    }

    expect(data.body.kind).to.equal("Deployment");

    await sleep(100); // for now let's sleep, we need better mechanism in the future

    let obj;
    try {
      obj = await k8sDynamicApi.read(deployObj, "true");
    } catch (error) {
      expect(JSON.stringify(error.body)).to.equal("");
    }

    expect(obj.body.status).not.to.be.undefined;
  });
});

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}
