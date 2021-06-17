const { assert } = require("chai");
const {
    lokiPortForward,
    queryLoki
} = require("./client");

async function checkLokiLogs(startTimestamp) {
    const cancelPortForward = lokiPortForward();

    const retries = 0;
    while (retries < 10) {
        const labels = '{app="commerce-mock", container="commerce-mock", namespace="mocks"}';
        const logs = await queryLoki(labels, startTimestamp);
        if (logs.streams.length > 0) {
            break;
        }
        console.log("retry num: ", retries);
        await sleep(1000);
        retries++;
    }
    assert.isNotEmpty(logs.streams, "No logs fetched from Loki");

    cancelPortForward();
}

module.exports = {
    checkLokiLogs
};
