const { assert } = require("chai");
const {
    lokiPortForward,
    queryLoki
} = require("./client");

async function checkLokiLogs(startTimestamp) {
    console.log("timestamp at which query is executed: ", new Date().toISOString());
    while(true) {

    }
    const cancelPortForward = lokiPortForward();

    const labels = '{app="commerce-mock", container="commerce-mock", namespace="mocks"}';
    const logs = await queryLoki(labels, startTimestamp);
    assert.isNotEmpty(logs.streams);

    cancelPortForward();
}

module.exports = {
    checkLokiLogs
};
