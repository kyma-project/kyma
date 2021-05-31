const { 
    waitForFunction,
    k8sApply,
    kubectlApplyDir,
    kubectlDeleteDir,
    sleep
} = require("../utils");
const { expect } = require("chai");


async function waitForK8sResources(){
    await waitForFunction("audit-test-fn", "audit-test", 120000);
}

async function createNamespace(namespace) {
    await k8sApply([namespaceObj(namespace)])
}

async function deployK8sResources() {
    await kubectlApplyDir("./audit-log/fixtures", "audit-test")
    await waitForK8sResources()
}

async function deleteK8sResources() {
    await kubectlDeleteDir("./audit-log/fixtures", "audit-test")
}

async function waitForAuditLogs() {
    await sleep(120000);
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
    var notFound = [];
    let groups = [
        { "metrics-foo": "monitoring.coreos.com" } ,
        { "audit-role-foo": "rbac.authorization.k8s.io"},
        {"audit-test-fn": "serverless.kyma-project.io"},
        {"foo-config": "configmaps"} // for checking configmap
    ]
    let actions = [
        "create",
        "delete"
    ]

    groups.forEach(group => {
        actions.forEach(action => {
            for (let resName in group) {
                let res = parseAuditLogs(logs, resName, group[resName],  action)
                if (res == false) {
                    notFound.push({
                        key: group[resName],
                        value: action
                    })
                }
            }
        });
    });
    if (notFound.length != 0) {
        notFound.forEach(el => {
            for (let key in el) {
                    console.log("Group: " + key + " with action: " + el[key] + "not found")
            }
        })
    }
    expect(notFound).to.be.empty
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
    createNamespace,
    deployK8sResources,
    deleteK8sResources,
    waitForAuditLogs,
    checkAuditLogs
}