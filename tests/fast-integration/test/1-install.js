const { installKyma } = require("../installer");

describe("Installation", function () {
  this.timeout(10 * 60 * 1000);

  it("Kyma should successfully install", async function () {
    const options = {
      skipComponents: ["dex","tracing","monitoring","console","kiali","logging"],
      newEventing: true
    };
    await installKyma(options);
  });
});
