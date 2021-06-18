const { installKyma } = require("../installer");

describe("Installation", function () {
  this.timeout(10 * 60 * 1000);

  it("Kyma should successfully install", async function () {
    const options = {
      skipComponents: ["dex","console"],
      newEventing: true,
      withCentralApplicationGateway: process.env.WITH_CENTRAL_APPLICATION_GATEWAY || false
    };
    await installKyma(options);
  });
});
