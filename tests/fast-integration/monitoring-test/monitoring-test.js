const uuid = require("uuid");
const axios = require('axios');

const {
  assert
} = require("chai");

const {
  debug,
  genRandom,
  kubectlPortForward,
} = require("../utils");

const {
  shouldIgnoreTarget
} = require('../monitoring/helpers')

var cancelPortForward;
let prometheusPort = 9090;

before(() => {
  cancelPortForward = kubectlPortForward("kyma-system", "prometheus-monitoring-prometheus-0", prometheusPort);
})

after(() => {
  cancelPortForward()
})

describe("Monitoring test", function () {
  const suffix = genRandom(4);
  const appName = `app-${suffix}`;
  const runtimeName = `kyma-${suffix}`;
  const scenarioName = `test-${suffix}`;
  const runtimeID = uuid.v4();

  debug(`RuntimeID ${runtimeID}`, `Scenario ${scenarioName}`, `Runtime ${runtimeName}`, `Application ${appName}`);

  this.timeout(60 * 60 * 1000 * 3); // 3h
  this.slow(5000);

  it("Should have all targets healthy", async () => {
    let response = await axios.get(`http://localhost:${prometheusPort}/api/v1/targets?state=active`);
    let responseBody = response.data;
    let activeTargets = responseBody.data.activeTargets;
    for (const target of activeTargets.filter(t => !shouldIgnoreTarget(t))) {
      assert.equal(target.health, "up")
    }
    console.log();
  });
});
