const { assert } = require("chai");
const { sleep, wait } = require ("../utils")

function findAuditLog(logs, group) {
    for (let element of logs) {
        if (element.message.includes(group.groupName)) {
            if (element.message.includes(group.resName)){
                if (element.message.includes(group.action)) {
                    return true
                }
            }
        }
    }
    return false;
}

function parseAuditLogs(logs, groups) {
    groups.forEach(group => {
        let found = false
        found = findAuditLog(logs, group)
        if (found == true) {
            const index = groups.indexOf(group);
            if (index > -1) {
                groups.splice(index, 1);
            }
        }
    })
    return groups
}

async function checkAuditLogs(cred, groups) {
    let retries = 0
    let notFound = groups
    while (retries < 15) {
        const logs = await cred.fetchLogs();
        assert.isNotEmpty(logs)
        notFound = parseAuditLogs(logs, notFound)
        await sleep(5*1000)
        retries++
    }
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