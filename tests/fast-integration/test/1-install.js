const { installKyma } = require("../installer");

describe("Installation", function () {
  this.timeout(10 * 60 * 1000);

  it("Kyma should successfully install", async function () {
    const options = {
      skipComponents: 'tracing,kiali,knative-eventing,knative-provisioner-natss,nats-streaming,event-sources',
      newEventing: true
    };
    await installKyma(options);
  });
});
