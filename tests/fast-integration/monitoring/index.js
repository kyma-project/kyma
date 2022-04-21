const prometheus = require('./prometheus');
const grafana = require('./grafana');
const {getEnvOrDefault, info} = require('../utils');

function monitoringTests() {
  if (getEnvOrDefault('KYMA_MAJOR_UPGRADE', 'false') === 'true') {
    info('Skipping monitoring tests for Kyma 1 to Kyma 2 upgrade scenario');
    return;
  }

  describe('Grafana Tests:', async function() {
    this.timeout(5 * 60 * 1000); // 5 min
    this.slow(5 * 1000);

    it('Grafana pods should be ready', async () => {
      await grafana.assertPodsExist();
    });

    it('Grafana redirects should work', async () => {
      await grafana.assertGrafanaRedirectsExist();
    });
  });

  describe('Prometheus Tests:', function() {
    this.timeout(5 * 60 * 1000); // 5 min
    this.slow(5000);

    it('Prometheus pods should be ready', async () => {
      await prometheus.assertPodsExist();
    });

    it('Prometheus targets should be healthy', async () => {
      await prometheus.assertAllTargetsAreHealthy();
    });

    it('No critical Prometheus alerts should be firing', async () => {
      await prometheus.assertNoCriticalAlertsExist();
    });

    it('Prometheus scrape pools should have a target', async () => {
      await prometheus.assertScrapePoolTargetsExist();
    });

    it('Prometheus rules should be registered', async () => {
      await prometheus.assertRulesAreRegistered();
    });

    it('Prometheus rules should be healthy', async () => {
      await prometheus.assertAllRulesAreHealthy();
    });

    it('Metrics used by Kyma Dashboard should exist', async () => {
      await prometheus.assertMetricsExist();
    });
  });
}

function setGrafanaProxy() {
  if (getEnvOrDefault('KYMA_MAJOR_UPGRADE', 'false') === 'true') {
    info('Skipping setting of Grafana Proxy for Kyma 1 to Kyma 2 upgrade scenario');
    return;
  }

  describe('Should prepare Grafana', function() {
    it('Setting Grafana Proxy', async () => {
      await grafana.setGrafanaProxy();
    });
  });
}

function resetGrafanaProxy() {
  if (getEnvOrDefault('KYMA_MAJOR_UPGRADE', 'false') === 'true') {
    info('Skipping resetting of Grafana Proxy for Kyma 1 to Kyma 2 upgrade scenario');
    return;
  }

  describe('Should ensure that Grafana is not exposed', function() {
    it('Resetting Grafana Proxy', async () => {
      await grafana.resetGrafanaProxy();
    });
  });
}

module.exports = {
  monitoringTests,
  setGrafanaProxy,
  resetGrafanaProxy,
};
