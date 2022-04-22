const {assert} = require('chai');

const {sleep} = require('../utils');
const {
  queryLoki,
} = require('./client');


async function checkLokiLogs(startTimestamp) {
  const labels = '{app="commerce-mock", container="mock", namespace="mocks"}';
  let logsLength = 0;
  for (let i = 0; i < 20; ++i) {
    const logs = await queryLoki(labels, startTimestamp);
    if (logs.streams.length > 0) {
      logsLength = logs.streams.length;
      break;
    }
    await sleep(5 * 1000);
  }
  assert.isAbove(logsLength, 0, 'No logs fetched from Loki');
}

module.exports = {
  checkLokiLogs,
};
