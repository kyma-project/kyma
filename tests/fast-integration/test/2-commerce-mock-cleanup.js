const {
  cleanMockTestFixture,
} = require("./fixtures/commerce-mock");
const {
} = require("../utils");

describe("CommerceMock cleanup", function () {
  this.timeout(10 * 60 * 1000);
  this.slow(5000);
  const testNamespace = "test";

  it("Test namespaces should be deleted", async function () {
    await cleanMockTestFixture("mocks", testNamespace, false);
  });

});
