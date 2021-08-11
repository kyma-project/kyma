const uuid = require("uuid");

const { assert } = require("chai");

const { waitForPodWithLabel } = require("../utils");

describe("Telemtry operator", function () {
  // check if operator installed
  it("Operator should be ready", async function () {
    let namespace = "kyma-system";
    await waitForPodWithLabel("control-plane", "controller-manager", namespace);
  });
});
