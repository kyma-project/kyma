const {
  ensureCommerceMockLocalTestFixture,
} = require("../test/fixtures/commerce-mock");
const {
} = require("../utils");

describe("Upgrade test preparation", function () {
  this.timeout(10 * 60 * 1000);
  this.slow(5000);
  const testNamespace = "test";

  it("CommerceMock test fixture should be ready", async function () {
    await ensureCommerceMockLocalTestFixture("mocks", testNamespace).catch((err) => {
      console.dir(err); // first error is logged
      return ensureCommerceMockLocalTestFixture("mocks", testNamespace);
    });
  });

});
