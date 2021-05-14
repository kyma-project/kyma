
const k8s = require("@kubernetes/client-node");
const { k8sApply, waitForFunction, waitForSubscription, genRandom, retryPromise, deleteNamespaces } = require("../utils")
const fs = require('fs')
const path = require('path');
const axios = require("axios");
const https = require("https");
const { expect } = require("chai");
const httpsAgent = new https.Agent({
  rejectUnauthorized: false, // curl -k
});
axios.defaults.httpsAgent = httpsAgent;

const lastorderFunctionYaml = fs.readFileSync(
  path.join(__dirname, "./fixtures/commerce-mock/lastorder-function.yaml"),
  {
    encoding: "utf8",
  }
);

const lastorderObjs = k8s.loadAllYaml(lastorderFunctionYaml);
const testNamespace = "test";


function eventingSubscription(eventType, sink, name, ns) {
  return {
    apiVersion: "eventing.kyma-project.io/v1alpha1",
    kind: "Subscription",
    metadata: {
      name,
      namespace: ns,
    },
    spec: {
      filter: {
        dialect: "beb",
        filters: [{
          eventSource: {
            property: "source", type: "exact", value: "",
          },
          eventType: {
            property: "type", type: "exact", value: eventType
          }
        }]
      },
      protocol: "BEB",
      protocolsettings: {
        exemptHandshake: true,
        qos: "AT-LEAST-ONCE",
      },
      sink
    }
  }
}


describe("In-cluster eventing with functions", function () {
  this.timeout(60 * 1000);

  it("function should be ready", async function () {
    await k8sApply([{
      apiVersion: "v1",
      kind: "Namespace",
      metadata: { name: testNamespace }
    }]);    
    await k8sApply(lastorderObjs, testNamespace, true);
    await waitForFunction("lastorder", testNamespace);
  });

  it("In-cluster event subscription should be ready", async function () {
    await k8sApply([eventingSubscription(
      `sap.kyma.custom.inapp.order.received.v1`,
      `http://lastorder.${testNamespace}.svc.cluster.local`,
      "lastorder",
      testNamespace)]);
    await waitForSubscription("lastorder", testNamespace);
  });

  it("in-cluster event should be delivered", async function () {
    const eventId = "event-"+genRandom(5);
    let response = await retryPromise(() => axios.post("https://lastorder.local.kyma.dev", { id: eventId }, {params:{send:true}}), 10, 1)
    response = await axios.get("https://lastorder.local.kyma.dev", { params: { inappevent: eventId } });
    console.dir(response.data);
    expect(response).to.have.nested.property("data.id", eventId, "The same event id expected in the result");
    expect(response).to.have.nested.property("data.shipped", true, "Order should have property shipped");

  });

  it("should clean up resources", function() {
    deleteNamespaces([testNamespace], true);
  })

});