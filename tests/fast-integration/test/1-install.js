const { installKyma } = require("../installer");
const { k8sCoreV1Api } = require("../utils");

describe("Installation", function () {
  this.timeout(10 * 60 * 1000);

  it("Kyma should successfully install", async function () {
    // temporary until kyma is provided via pipeline
    const result = await k8sCoreV1Api.listNamespace();
    if (result && result.body.items.map((i) => i.metadata.name).includes('kyma-system')) {
      console.log("Namespace 'kyma-system' exists. Skipping installation.");
      return;
    }

    const options = {
      skipComponents: ["dex","console"],
      withCentralAppConnectivity: (process.env.WITH_CENTRAL_APP_CONNECTIVITY === "true")
    };
    await installKyma(options);
  });
});
