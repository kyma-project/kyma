const prometheus = require('../monitoring/prometheus');
const grafana = require('../monitoring/grafana');

const {prometheusPortForward} = require('../monitoring/client');

function monitoringTests() {
  describe('Prometheus Tests:', function() {
    this.timeout(5 * 60 * 1000); // 5 min
    this.slow(5 * 1000);

    let cancelPortForward;

    before(async () => {
      cancelPortForward = prometheusPortForward();
    });

    after(async () => {
      cancelPortForward();
    });

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
}

module.exports = {
  monitoringTests,
};
