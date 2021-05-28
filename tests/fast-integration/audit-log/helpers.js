const { 
    waitForVirtualService,
    waitForFunction,
    k8sApply
} = require("../utils");

async function waitForK8sResources(){
    await waitForFunction("audit-test-fn", "audit-test");
}

async function createNamespace(namespace) {
    await k8sApply([namespaceObj(namespace)])
}


function namespaceObj(name) {
    return {
      apiVersion: "v1",
      kind: "Namespace",
      metadata: { name },
    };
  }

module.exports = {
    waitForK8sResources,
    createNamespace
}