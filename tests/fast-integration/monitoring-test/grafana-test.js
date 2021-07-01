const {
    assert
  } = require("chai");
  const {
    assertGrafanaredirect,
  } = require("../monitoring/helpers");  

  describe("Grafana test", function () {
    this.timeout(30 * 60 * 1000); // 30 min
    this.slow(5 * 1000);

    it("Checks grafana redirect", async () => {
        assertGrafanaredirect()
    })
  })
