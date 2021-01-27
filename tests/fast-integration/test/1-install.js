const { installKyma } = require("../installer");

describe("Installation", function () {
  this.timeout(10 * 60 * 1000);

  it("Kyma should successfully install", async function () {
    await installKyma(undefined, "1.7.6");
  });
});
