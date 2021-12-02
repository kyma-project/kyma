const {
    checkGrafanaRedirects
} = require("./../monitoring/grafana");

function observabilityTests() {
    // describe("Monitoring tests", function () {
    //     this.timeout(30 * 60 * 1000); // 30 min
    //     this.slow(5 * 1000);

    //     var cancelPortForward;

    //     before(async () => {
    //       cancelPortForward = prometheusPortForward();
    //     });


    //     after(async () => {
    //       cancelPortForward();
    //     });
    // })

    describe("Grafana tests", async function () {
        this.timeout(30 * 60 * 1000); // 30 min
        this.slow(5 * 1000);

        it("Checking Grafana redirects", async () => {
            await checkGrafanaRedirects();
        });
    })
}

module.exports = {
    observabilityTests,
}