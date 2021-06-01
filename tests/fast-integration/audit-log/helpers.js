const { 
    waitForFunction,
    k8sApply,
    kubectlApplyDir,
    kubectlDeleteDir,
    sleep
} = require("../utils");
const { assert } = require("chai");

const testNS = "audit-test"
const fnName = "audit-test-fn"

async function createNamespace() {
    await k8sApply([namespaceObj(testNS)])
}


async function deployK8sResources() {
    await kubectlApplyDir("./audit-log/fixtures", testNS)
    await waitForFunction(fnName, testNS, 120000);

}

async function deleteK8sResources() {
    await kubectlDeleteDir("./audit-log/fixtures", "audit-test")
}

async function waitForAuditLogs() {
    await sleep(60000);
}

function parseAuditLogs(logs, resName, groupName, action) {
    let found = new Boolean(false)
    logs.forEach(element => {
        if (element.message.includes(groupName)) {
            if (element.message.includes(resName)){
                if (element.message.includes(action)) {
                    found = true
                }
            }
        }
    });
    return found
}

async function checkAuditLogs(cred) {
    let logs = await cred.fetchLogs();
    assert.isNotEmpty(logs)
    var notFound = [];
    const groups = [
        { "metrics-foo": "monitoring.coreos.com" } ,
        { "audit-role-foo": "rbac.authorization.k8s.io"},
        {"audit-test-fn": "serverless.kyma-project.io"},
        {"foo-config": "configmaps"} // for checking configmap
    ]
    const actions = [
        "create",
        "delete"
    ]

    groups.forEach(group => {
        actions.forEach(action => {
            for (let resName in group) {
                let res = parseAuditLogs(logs, resName, group[resName],  action)
                if (res == false) {
                    let resNotfound = new Map()
                    resNotfound.set(group[resName],action )
                    notFound.push(resNotfound)
                }
            }
        });
    });
    if (notFound.length != 0) {
        notFound.forEach(el => {
            console.log("Following groups and actions not found: " , el)
        })
    }
    assert.isEmpty(notFound, `Number of groups not found to be zero`)
}

function namespaceObj(name) {
    return {
      apiVersion: "v1",
      kind: "Namespace",
      metadata: { name },
    };
  }

module.exports = {
    createNamespace,
    deployK8sResources,
    deleteK8sResources,
    waitForAuditLogs,
    checkAuditLogs
}