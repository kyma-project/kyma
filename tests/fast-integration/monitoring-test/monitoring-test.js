const uuid = require("uuid");
const axios = require('axios');

const {
  debug,
  genRandom,
  kubectlPortForward,
} = require("../utils");

describe("Monitoring test", function () {

  const suffix = genRandom(4);
  const appName = `app-${suffix}`;
  const runtimeName = `kyma-${suffix}`;
  const scenarioName = `test-${suffix}`;
  const runtimeID = uuid.v4();


  debug(`RuntimeID ${runtimeID}`, `Scenario ${scenarioName}`, `Runtime ${runtimeName}`, `Application ${appName}`);

  const testNS = "monitoring-test";

  this.timeout(60 * 60 * 1000 * 3); // 3h
  this.slow(5000);

  //it("should have all Rules healthy", async function() {
  //  await checkPrometheusRules();
  //});

  it("Listing all pods in cluster", async function () {
    let prometheusPort = 9090;
    let cleanup = kubectlPortForward("kyma-system", "prometheus-monitoring-prometheus-0", prometheusPort);
    let response = await axios.get(`http://localhost:${prometheusPort}/api/v1/targets?state=active`);
    console.log(JSON.stringify(response.data));
    cleanup();
  });
});
