const {
  updateSKR,
  ensureValidShootOIDCConfig,
  ensureValidOIDCConfigInCustomerFacingKubeconfig,
} = require('../kyma-environment-broker');
const {
  ensureCommerceMockWithCompassTestFixture,
  checkFunctionResponse,
  sendLegacyEventAndCheckResponse,
  deleteMockTestFixture, sendCloudEventStructuredModeAndCheckResponse, sendCloudEventBinaryModeAndCheckResponse,
  checkInClusterEventDelivery,
} = require('../test/fixtures/commerce-mock');
const {
  ensureKymaAdminBindingExistsForUser,
  ensureKymaAdminBindingDoesNotExistsForUser,
} = require('../utils');
const {
  AuditLogCreds,
  AuditLogClient,
  checkAuditLogs,
} = require('../audit-log');
const {keb, gardener, director} = require('./helpers');
const {KCPWrapper, KCPConfig} = require('../kcp/client');

const kcp = new KCPWrapper(KCPConfig.fromEnv());

const updateTimeout = 1000 * 60 * 20; // 20m

function oidcE2ETest() {
  describe('OIDC E2E Test', function() {
    it('Assure initial OIDC config is applied on shoot cluster', async function() {
      ensureValidShootOIDCConfig(this.shoot, this.options.oidc0);
    });

    it('Assure initial OIDC config is part of kubeconfig', async function() {
      await ensureValidOIDCConfigInCustomerFacingKubeconfig(keb, this.options.instanceID, this.options.oidc0);
    });

    it('Assure initial cluster admin', async function() {
      await ensureKymaAdminBindingExistsForUser(this.options.administrator0);
    });

    it('Update SKR service instance with OIDC config', async function() {
      const customParams = {
        oidc: this.options.oidc1,
      };
      const skr = await updateSKR(keb,
          kcp,
          gardener,
          this.options.instanceID,
          this.shoot.name,
          customParams,
          updateTimeout,
          null,
          false);
      this.shoot = skr.shoot;
    });

    it('Should get Runtime Status after updating OIDC config', async function() {
      try {
        const runtimeStatus = await kcp.getRuntimeStatusOperations(this.options.instanceID);
        console.log(`\nRuntime status: ${runtimeStatus}`);
        await kcp.reconcileInformationLog(runtimeStatus);
      } catch (e) {
        console.log(`before hook failed: ${e.toString()}`);
      }
    });

    it('Assure updated OIDC config is applied on shoot cluster', async function() {
      ensureValidShootOIDCConfig(this.shoot, this.options.oidc1);
    });

    it('Assure updated OIDC config is part of kubeconfig', async function() {
      await ensureValidOIDCConfigInCustomerFacingKubeconfig(keb, this.options.instanceID, this.options.oidc1);
    });

    it('Assure cluster admin is preserved', async function() {
      await ensureKymaAdminBindingExistsForUser(this.options.administrator0);
    });

    it('Update SKR service instance with new admins', async function() {
      const customParams = {
        administrators: this.options.administrators1,
      };
      const skr = await updateSKR(keb,
          kcp,
          gardener,
          this.options.instanceID,
          this.shoot.name,
          customParams,
          updateTimeout,
          null,
          false);
      this.shoot = skr.shoot;
    });

    it('Should get Runtime Status after updating admins', async function() {
      const runtimeStatus = await kcp.getRuntimeStatusOperations(this.options.instanceID);
      console.log(`\nRuntime status: ${runtimeStatus}`);
      await kcp.reconcileInformationLog(runtimeStatus);
    });

    it('Assure only new cluster admins are configured', async function() {
      await ensureKymaAdminBindingExistsForUser(this.options.administrators1[0]);
      await ensureKymaAdminBindingExistsForUser(this.options.administrators1[1]);
      await ensureKymaAdminBindingDoesNotExistsForUser(this.options.administrator0);
    });
  });
}

function commerceMockTest() {
  describe('CommerceMockTest()', function() {
    const AWS_PLAN_ID = '361c511f-f939-4621-b228-d0fb79a1fe15';

    it('CommerceMock test fixture should be ready', async function() {
      await ensureCommerceMockWithCompassTestFixture(
          director,
          this.options.appName,
          this.options.scenarioName,
          'mocks',
          this.options.testNS,
          true,
      );
    });

    it('in-cluster event should be delivered (structured and binary mode)', async function() {
      await checkInClusterEventDelivery(this.options.testNS);
    });

    it('function should be reachable through secured API Rule', async function() {
      await checkFunctionResponse(this.options.testNS);
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

    it('Deletes the resources that have been created', async function() {
      await deleteMockTestFixture('mocks', this.options.testNS);
    });

    // Check audit log for AWS
    if (process.env.KEB_PLAN_ID === AWS_PLAN_ID) {
      const auditlogs = new AuditLogClient(AuditLogCreds.fromEnv());

      it('Check audit logs', async function() {
        await checkAuditLogs(auditlogs, null);
      });

      // TODO: Enable checkAuditEventsThreshold again when fix is ready by Andreas Thaler

      // it('Amount of audit events must not exceed a certain threshold', async function() {
      //   await checkAuditEventsThreshold(4);
      // });
    }
  });
}

module.exports = {
  commerceMockTest,
  oidcE2ETest,
};
