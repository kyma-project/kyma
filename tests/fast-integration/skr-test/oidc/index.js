const {expect} = require('chai');
const {
  updateSKR,
  ensureValidShootOIDCConfig,
  ensureValidOIDCConfigInCustomerFacingKubeconfig,
} = require('../../kyma-environment-broker');
const {
  ensureKymaAdminBindingExistsForUser,
  ensureKymaAdminBindingDoesNotExistsForUser,
} = require('../../utils');
const {keb, kcp, gardener} = require('../helpers');

const updateTimeout = 1000 * 60 * 20; // 20m

function oidcE2ETest(options, getShootInfoFunc) {
  describe('OIDC Test', function() {
    let shoot = undefined;
    let givenOidcConfig = undefined;

    before('Get provisioned Shoot Info', async function() {
      shoot = getShootInfoFunc();
      expect(shoot).to.not.be.undefined;
      givenOidcConfig = shoot.oidcConfig;
    });

    it('Assure initial OIDC config is applied on shoot cluster', async function() {
      ensureValidShootOIDCConfig(shoot, givenOidcConfig);
    });

    it('Assure initial OIDC config is part of kubeconfig', async function() {
      await ensureValidOIDCConfigInCustomerFacingKubeconfig(keb, options.instanceID, givenOidcConfig);
    });

    it('Assure initial cluster admin', async function() {
      await ensureKymaAdminBindingExistsForUser(options.kebUserId); // default user id
    });

    it('Update SKR service instance with OIDC config', async function() {
      this.timeout(updateTimeout);
      const customParams = {
        oidc: options.oidc1,
      };
      const skr = await updateSKR(keb,
          kcp,
          gardener,
          options.instanceID,
          shoot.name,
          customParams,
          updateTimeout,
          null,
          false);
      shoot = skr.shoot;
    });

    it('Should get Runtime Status after updating OIDC config', async function() {
      try {
        const runtimeStatus = await kcp.getRuntimeStatusOperations(options.instanceID);
        console.log(`\nRuntime status: ${runtimeStatus}`);
        await kcp.reconcileInformationLog(runtimeStatus);
      } catch (e) {
        console.log(`before hook failed: ${e.toString()}`);
      }
    });

    it('Assure updated OIDC config is applied on shoot cluster', async function() {
      ensureValidShootOIDCConfig(shoot, options.oidc1);
    });

    it('Assure updated OIDC config is part of kubeconfig', async function() {
      await ensureValidOIDCConfigInCustomerFacingKubeconfig(keb, options.instanceID, options.oidc1);
    });

    it('Assure cluster admin is preserved', async function() {
      await ensureKymaAdminBindingExistsForUser(options.kebUserId);
    });

    it('Update SKR service instance with new admins', async function() {
      this.timeout(updateTimeout);
      const customParams = {
        administrators: options.administrators1,
      };
      const skr = await updateSKR(keb,
          kcp,
          gardener,
          options.instanceID,
          shoot.name,
          customParams,
          updateTimeout,
          null,
          false);

      shoot = skr.shoot;
    });

    it('Should get Runtime Status after updating admins', async function() {
      const runtimeStatus = await kcp.getRuntimeStatusOperations(options.instanceID);
      console.log(`\nRuntime status: ${runtimeStatus}`);
      await kcp.reconcileInformationLog(runtimeStatus);
    });

    it('Assure only new cluster admins are configured', async function() {
      await ensureKymaAdminBindingExistsForUser(options.administrators1[0]);
      await ensureKymaAdminBindingExistsForUser(options.administrators1[1]);
      await ensureKymaAdminBindingDoesNotExistsForUser(options.kebUserId);
    });

    it('Update SKR service instance with initial OIDC config and admins', async function() {
      this.timeout(updateTimeout);
      const customParams = {
        oidc: givenOidcConfig,
        administrators: options.kebUserId,
      };
      const skr = await updateSKR(keb,
          kcp,
          gardener,
          options.instanceID,
          shoot.name,
          customParams,
          updateTimeout,
          null,
          false);

      shoot = skr.shoot;
    });

    it('Should get Runtime Status after updating OIDC config and admins', async function() {
      const runtimeStatus = await kcp.getRuntimeStatusOperations(options.instanceID);
      console.log(`\nRuntime status: ${runtimeStatus}`);
      await kcp.reconcileInformationLog(runtimeStatus);
    });

    it('Assure updated (with initial values) OIDC config is applied on shoot cluster', async function() {
      ensureValidShootOIDCConfig(shoot, givenOidcConfig);
    });

    it('Assure updated (with initial values) OIDC config is part of kubeconfig', async function() {
      await ensureValidOIDCConfigInCustomerFacingKubeconfig(keb, options.instanceID, givenOidcConfig);
    });

    it('Assure only initial cluster admins are configured', async function() {
      await ensureKymaAdminBindingDoesNotExistsForUser(options.administrators1[0]);
      await ensureKymaAdminBindingDoesNotExistsForUser(options.administrators1[1]);
      await ensureKymaAdminBindingExistsForUser(options.kebUserId);
    });
  });
}

module.exports = {
  oidcE2ETest,
};
