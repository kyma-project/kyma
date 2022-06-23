const {
  getContainerRestartsForAllNamespaces,
  printRestartReport,
} = require('../../utils');
const {
  checkFunctionResponse,
  checkInClusterEventDelivery,
  sendLegacyEventAndCheckResponse,
  sendCloudEventStructuredModeAndCheckResponse,
  sendCloudEventBinaryModeAndCheckResponse,
  deleteMockTestFixture,
  ensureCommerceMockWithCompassTestFixture,
} = require('../../test/fixtures/commerce-mock');
const {
  AuditLogCreds,
  AuditLogClient,
  checkAuditLogs,
} = require('../../audit-log');
const {debug} = require('../../utils');
const {director} = require('../helpers');
const {
  exposeGrafana,
  unexposeGrafana,
} = require('../../monitoring');

const AWS_PLAN_ID = '361c511f-f939-4621-b228-d0fb79a1fe15';
// rate of audit events generated in 60 minutes. The value is derived from the actual prometheus query.
// const auditLogsThreshold = 6;
const testTimeout = 1000 * 60 * 30; // 30m

// prepares all the resources required for commerce mock to be executed;
// runs the actual tests and checks the audit logs in case of AWS plan
function commerceMockTest(options) {
  describe('CommerceMock Test', function() {
    this.timeout(testTimeout);
    commerceMockTestPreparation(options);
    commerceMockTests(options.testNS);
    commerceMockCleanup(options.testNS);

    context('Check audit logs for AWS', function() {
      if (process.env.KEB_PLAN_ID === AWS_PLAN_ID) {
        checkAuditLogsForAWS();
      } else {
        debug('Skipping step for non-AWS plan');
      }
    });
  });
}


function commerceMockTestPreparation(options) {
  it('CommerceMock test fixture should be ready', async function() {
    await ensureCommerceMockWithCompassTestFixture(
        director,
        options.appName,
        options.scenarioName,
        'mocks',
        options.testNS,
        true,
    );
  });
}

// executes the actual commerce mock tests
function commerceMockTests(testNamespace) {
  let initialRestarts = undefined;

  it('Listing all pods in cluster', async function() {
    initialRestarts = await getContainerRestartsForAllNamespaces();
  });

  it('in-cluster event should be delivered (structured and binary mode)', async function() {
    await checkInClusterEventDelivery(testNamespace);
  });

  it('function should be reachable through secured API Rule', async function() {
    await checkFunctionResponse(testNamespace);
  });

  it('order.created.v1 legacy event should trigger the lastorder function', async function() {
    await sendLegacyEventAndCheckResponse();
  });

  it('order.created.v1 cloud event in structured mode should trigger the lastorder function', async function() {
    await sendCloudEventStructuredModeAndCheckResponse();
  });

  it('order.created.v1 cloud event in binary mode should trigger the lastorder function', async function() {
    await sendCloudEventBinaryModeAndCheckResponse();
  });

  it('Should print report of restarted containers, skipped if no crashes happened', async function() {
    const afterTestRestarts = await getContainerRestartsForAllNamespaces();
    printRestartReport(initialRestarts, afterTestRestarts);
  });
}

function commerceMockCleanup(testNamespace) {
  it('CommerceMock test fixture should be deleted', async function() {
    await deleteMockTestFixture('mocks', testNamespace);
  });
}

function checkAuditLogsForAWS() {
  // it('Expose Grafana', async function() {
  //   await exposeGrafana();
  // });

  it('Check audit logs', async function() {
    const auditLogs = new AuditLogClient(AuditLogCreds.fromEnv());
    await checkAuditLogs(auditLogs, null);
  });

  // it('Amount of audit events must not exceed a certain threshold', async function() {
  //   await checkAuditEventsThreshold(auditLogsThreshold);
  // });
  //
  // it('Unexpose Grafana', async function() {
  //   await unexposeGrafana(true);
  // });
}

module.exports = {
  commerceMockTest,
};
