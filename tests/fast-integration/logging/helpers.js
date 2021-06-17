const { assert } = require("chai");

const { sleep } = require ("../utils");
const {
    lokiPortForward,
    queryLoki
} = require("./client");

async function checkLokiLogs(startTimestamp) {
    const cancelPortForward = lokiPortForward();
    const labels = '{app="commerce-mock", container="commerce-mock", namespace="mocks"}';
    let logsFetched = false;
    let retries = 0;
    while (retries < 10) {
        const logs = await queryLoki(labels, startTimestamp);
        if (logs.streams.length > 0) {
            logsFetched = true;
            break;
        }
        console.log("retry num: ", retries);
        await sleep(1000);
        retries++;
    }
    assert.isTrue(logsFetched, "No logs fetched from Loki");
    cancelPortForward();
}

module.exports = {
    checkLokiLogs
};
