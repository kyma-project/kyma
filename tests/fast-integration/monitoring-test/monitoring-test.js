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
  shouldIgnoreTarget,
  shouldIgnoreAlert,
  buildScrapePoolSet,
} = require('../monitoring/helpers')

describe("Monitoring test", function () {
  
  const suffix = genRandom(4);
  const appName = `app-${suffix}`;
  const runtimeName = `kyma-${suffix}`;
  const scenarioName = `test-${suffix}`;
  const runtimeID = uuid.v4();

  debug(`RuntimeID ${runtimeID}`, `Scenario ${scenarioName}`, `Runtime ${runtimeName}`, `Application ${appName}`);

  this.timeout(60 * 60 * 1000 * 3); // 3h
  this.slow(5000);

  var cancelPortForward;
  let prometheusPort = 9090;

  before(() => {
    cancelPortForward = kubectlPortForward("kyma-system", "prometheus-monitoring-prometheus-0", prometheusPort);
  })
  
  after(() => {
    cancelPortForward()
  })

  it("All targets should be healthy", async () => {
    let response = await axios.get(`http://localhost:${prometheusPort}/api/v1/targets?state=active`);
    let responseBody = response.data;
    let activeTargets = responseBody.data.activeTargets;
    let unhealthyTargets = activeTargets.filter(t => !shouldIgnoreTarget(t) && t.health != "up").map(t => t.discoveredLabels.job);
    
    assert.isEmpty(unhealthyTargets, `Following targets are unhealthy: ${unhealthyTargets.join(", ")}`);
  });

  it("There should be no firing critical alerts", async () => {
    let response = await axios.get(`http://localhost:${prometheusPort}/api/v1/alerts`);
    let responseBody = response.data;
    let allAlerts = responseBody.data.alerts;
    let firingAlerts = allAlerts.filter(a => !shouldIgnoreAlert(a) && a.state == 'firing').map(a => a.labels.alertname);
    
    assert.isEmpty(firingAlerts, `Following alerts are firing: ${firingAlerts.join(", ")}`);
  });

  it("All pods should be ready", async () => {
    //TODO
  });

  it("Each scrape pool should have a healthy target", async () => {
    //TODO
    let scrapePools = await buildScrapePoolSet();

    let response = await axios.get(`http://localhost:${prometheusPort}/api/v1/targets?state=active`);
    let responseBody = response.data;
    let activeTargets = responseBody.data.activeTargets;

    for (const target of activeTargets) {
      scrapePools.delete(target.scrapePool);
    }
    assert.isEmpty(scrapePools, `Following scrape pools have no targets: ${scrapePools}`)
  });

  it("All rules should be healthy", async () => {
    //TODO
  });
  
  it("Grafana should be ready", async () => {
    //TODO
  });
  
  it("Lambda UI dashboard should be ready", async () => {
    //TODO
  });

});
