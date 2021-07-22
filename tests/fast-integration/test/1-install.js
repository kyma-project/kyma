const { installKyma } = require("../installer");

describe("Installation", function () {
  this.timeout(10 * 60 * 1000);

  it("Kyma should successfully install", async function () {
    const options = {
      skipComponents: ["dex","console"],
      newEventing: true,
      centralApplicationConnectivity: (process.env.CENTRAL_APPLICATION_CONNECTIVITY === "true")
    };
    await installKyma(options);
  });
});
