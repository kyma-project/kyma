const { assert } = require("chai");

function parseAuditLogs(logs, resName, groupName, action) {
    logs.forEach(element => {
        if (element.message.includes(groupName)) {
            if (element.message.includes(resName)){
                if (element.message.includes(action)) {
                    return true
                }
            }
        }
    });
    return false
}

async function checkAuditLogs(cred, groups, actions) {
    const logs = await cred.fetchLogs();
    assert.isNotEmpty(logs)
    const notFound = [];

    groups.forEach(group => {
        actions.forEach(action => {
            for (let resName in group) {
                let res = parseAuditLogs(logs, resName, group[resName],  action)
                if (res == false) {
                    let resNotfound = new Map()
                    resNotfound.set(group[resName],action)
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

module.exports = {
    checkAuditLogs
}