const prometheus = require("./../monitoring/prometheus");
const grafana = require("./../monitoring/grafana");

const { prometheusPortForward } = require("../monitoring/client");

export function observabilityTests() {

    describe("Prometheus tests", function () {
        this.timeout(30 * 60 * 1000); // 30 min
        this.slow(5 * 1000);

        var cancelPortForward;

        before(async () => {
            cancelPortForward = prometheusPortForward();
        });

        after(async () => {
            cancelPortForward();
        });

        it("Prometheus pods should be ready", async () => {
            await prometheus.assertPodsExist();
        });

        it("Prometheus targets should be healthy", async () => {
            await prometheus.assertAllTargetsAreHealthy();
        });

        it("No critical Prometheus alerts should be firing", async () => {
            await prometheus.assertNoCriticalAlertsExist();
        });

        it("Prometheus scrape pools should have a target", async () => {
            await prometheus.assertScrapePoolTargetsExist();
        });

        it("Prometheus rules should be registered", async () => {
            await prometheus.assertRulesAreRegistered();
        });

        it("Prometheus rules should be healthy", async () => {
            await prometheus.assertAllRulesAreHealthy();
        });

        it("Metrics used by Kyma Dashboard should exist", async () => {
            await prometheus.assertMetricsExist();
        });
    });

    describe("Grafana tests", async function () {
        this.timeout(5 * 60 * 1000); // 5 min
        this.slow(5 * 1000);

        it("Grafana pods should be ready", async () => {
            await grafana.assertPodsExist();
        });

        it("Grafana redirects should work", async () => {
            await grafana.assertGrafanaRedirectsExist();
        });
    });
}
