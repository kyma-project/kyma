const { expect } = require("chai");
const { retryPromise } = require("./");

describe("retryPromise", () => {
  it("should pass original resolve value", async () => {
    const resolveMsg = "always resolves";
    const fn = () => Promise.resolve(resolveMsg);
    const val = await retryPromise(fn, 1, 10);
    expect(val).to.be.string(resolveMsg);
  });
  it("should retry several times and return if original promise resolved", async () => {
    let count = 0;

    const fn = () =>
      new Promise((resolve, reject) => {
        if (count < 5) {
          count++;
          reject("Count is too low!");
        }
        resolve("pass");
      });

    return retryPromise(fn, 50, 10);
  });
  it("should fail if promise rejects during all retries", async () => {
    const rejectReason = "always rejects";
    const fn = () => Promise.reject(rejectReason);

    try {
      await retryPromise(fn, 3, 10);
    } catch (err) {
      expect(err).to.equal(rejectReason);
    }
  });

  it("should fail on 0 retries and original promise reject", () => {
    try {
      retryPromise(() => {}, 0, 10); // 50 rrtures
    } catch (err) {
      expect(err.message).to.include("greater then 0");
    }
  });
});
