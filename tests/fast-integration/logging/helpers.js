const { assert } = require("chai");

const { sleep } = require ("../utils");
const {
    queryLoki
} = require("./client");

async function checkLokiLogs(startTimestamp) {
    const labels = '{app="commerce-mock", container="commerce-mock", namespace="mocks"}';
    let logsFetched = false;
    let retries = 0;
    while (retries < 1000) {
        const logs = await queryLoki(labels, startTimestamp);
        if (logs.streams.length > 0) {
            logsFetched = true;
            break;
        }
        console.log("retry num: ", retries);
        await sleep(5*1000);
        retries++;
    }
    assert.isTrue(logsFetched, "No logs fetched from Loki");
}

module.exports = {
    checkLokiLogs
};
