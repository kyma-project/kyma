const { ensureGettingStartedTestFixture,
  verifyOrderPersisted,
  cleanGettingStartedTestFixture
} = require('./fixtures/getting-started-guides')

describe("Getting Started Guide Tests", function () {
  this.timeout(3 * 60 * 1000);
  this.slow(5000);

  it("Getting started guide fixture should be ready", async function () {
    this.timeout(4 * 60 * 1000);
    await ensureGettingStartedTestFixture()
  });

  it("Order should be persisted and should survive pod restarts (redis storage)", async function () {
    await verifyOrderPersisted();
  })

  it("Namespace should be deleted", async function () {
    this.timeout(60 * 1000);
    await cleanGettingStartedTestFixture(false);
  })
});

