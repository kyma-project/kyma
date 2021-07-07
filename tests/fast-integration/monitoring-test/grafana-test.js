const {
    assert
  } = require("chai");
  const {
    assertGrafanaredirect,
    updateProxyDeployment,
    createSecret,
    restartProxyPod,
  } = require("../monitoring/helpers");  

  describe("Grafana test", function () {
    this.timeout(30 * 60 * 1000); // 30 min
    this.slow(5 * 1000);

    if (process.env.isKyma2 === "true") {
      it("Checks grafana redirect to kyma docs", async () => {
        let res = await assertGrafanaredirect("https://kyma-project.io/docs")
        assert.isTrue(res, "Grafana redirect to kyma docs does not work!");
      })

      it ("Creates secret for atuh proxy redirect", async() => {
        await createSecret()
      })

      it ("Restarts the proxy pod", async() => {
        await restartProxyPod()
      })

      it("Checks grafana redirect to OIDC provider", async() => {
        let res = await assertGrafanaredirect("https://accounts.google.com/signin/oauth")
        assert.isTrue(res, "Grafana redirect to google does not work!");
      })

      it("Updates proxy deployment", async() => {
        await updateProxyDeployment()
      })

      it("Checks authentication works and redirects to grafana URL", async() => {
        let res = await assertGrafanaredirect("https://grafana.")
        assert.isTrue(res, "Grafana redirect to grafana landing page does not work!");
      })

    } else {
      it("Checks grafana redirect to Dex", async () => {
        let res = await assertGrafanaredirect("https://dex.")
        assert.isTrue(res, "Grafana redirect to dex does not work!");
        })

    }
  })