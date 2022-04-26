const {assert} = require('chai');
const {sleep} = require('../utils');
const {queryPrometheus} = require('../monitoring/client');

function findAuditLog(logs, group) {
  for (const element of logs) {
    if (element.message.includes(group.groupName) &&
          element.message.includes(group.resName) &&
          element.message.includes(group.action)) {
      return true;
    }
  }
  return false;
}

function parseAuditLogs(logs, groups) {
  groups.forEach((group) => {
    const found = findAuditLog(logs, group);
    if (found === true) {
      const index = groups.indexOf(group);
      if (index > -1) {
        groups.splice(index, 1);
      }
    }
  });
  return groups;
}

async function checkAuditLogs(client, groups) {
  let retries = 0;
  let notFound = [
    {'resName': 'lastorder', 'groupName': 'serverless.kyma-project.io', 'action': 'create'},
    {'resName': 'lastorder', 'groupName': 'serverless.kyma-project.io', 'action': 'delete'},
    {'resName': 'commerce-mock', 'groupName': 'deployments', 'action': 'create'},
    {'resName': 'commerce-mock', 'groupName': 'deployments', 'action': 'delete'},
  ];

  while (retries < 15) {
    const logs = await client.fetchLogs();
    assert.isNotEmpty(logs);
    notFound = parseAuditLogs(logs, notFound);
    await sleep(5*1000);
    retries++;
  }
  if (notFound.length > 0) {
    notFound.forEach((el) => {
      console.log('Following groups and actions not found: ', el);
    });
  }
  assert.isEmpty(notFound, `Number of groups not found to be zero`);
}

async function checkAuditEventsThreshold(threshold) {
  // Get the max rate for apiserver audit events over the last 60 min
  const query = 'max_over_time(rate(apiserver_audit_event_total{job="apiserver"}[1m])[60m:])';
  const result = await queryPrometheus(query);
  assert.isNotEmpty(result, 'The metrics "apiserver_audit_event_total" should not be empty! ');
  const maxAuditEventsRate = result[0].value[1];
  assert.isBelow(parseFloat(maxAuditEventsRate), threshold);
}

module.exports = {
  checkAuditLogs,
  checkAuditEventsThreshold,
};
