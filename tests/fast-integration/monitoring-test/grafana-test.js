const {
  checkGrafanaRedirectsInKyma1,
  checkGrafanaRedirectsInKyma2
} = require("../monitoring/helpers"); 
const {
  getEnvOrThrow
} = require("../utils");


describe("Grafana test", async function () {
  this.timeout(30 * 60 * 1000); // 30 min
  this.slow(5 * 1000);

  it("Checking Grafana redirects", async () => {
    const isKymaAlpha = getEnvOrThrow("KYMA_ALPHA");
    if (isKymaAlpha === "true") {
      await checkGrafanaRedirectsInKyma2();
    } else {
      await checkGrafanaRedirectsInKyma1();
    }
  });
})
