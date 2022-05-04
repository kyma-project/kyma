const {gatherOptions, withInstanceID, delay, waitForReconciliation} = require('./');
const {provisionSKR, deprovisionSKR} = require('../kyma-environment-broker');
const {keb, gardener, director} = require('./helpers');
const {initializeK8sClient} = require('../utils');
const {unregisterScenarioFromCompass, addScenarioInCompass, assignRuntimeToScenario, unregisterRuntimeFromCompass} = require('../compass');
const {oidcE2ETest, commerceMockTest} = require('./skr-test-N');
const {KCPWrapper, KCPConfig} = require('../kcp/client');

const N = process.env.N_TIMES > 1 ? process.env.N_TIMES : 1;
const kcp = new KCPWrapper(KCPConfig.fromEnv());

describe('Execute SKR test',  async function () {
  this.timeout(60 * 60 * 1000 * 3); // 3h
  this.slow(5000);

  const provisioningTimeout = 1000 * 60 * 30; // 30m
  const deprovisioningTimeout = 1000 * 60 * 95; // 95m

  before('Provision SKR', async function () {
    try {
      this.options = gatherOptions();
      console.log(`Provision SKR with instance ID ${this.options.instanceID}`);
      const customParams = {
        oidc: this.options.oidc0,
      };
      const skr = await provisionSKR(keb, kcp, gardener,
          this.options.instanceID,
          this.options.runtimeName,
          null,
          null,
          customParams,
          provisioningTimeout);

      this.shoot = skr.shoot;

      const runtimeStatus = await kcp.getRuntimeStatusOperations(this.options.instanceID);
      console.log(`\nRuntime status after provisioning: ${runtimeStatus}`);
      initializeK8sClient({kubeconfig: this.shoot.kubeconfig});
    } catch (e) {
      throw new Error(`before hook failed: ${e.toString()}`);
    } finally {
      const runtimeStatus = await kcp.getRuntimeStatusOperations(this.options.instanceID);
      await kcp.reconcileInformationLog(runtimeStatus);
    }
  });

  for (let i = 0; i < N; i++) {
    describe(`Loop ${i + 1}`, function () {
      before("Add scenario and assign runtime", async function () {
        const version = await kcp.version([]);
        console.log(version);

        this.options = gatherOptions(
            withInstanceID(this.options.instanceID),
        );
        await addScenarioInCompass(director, this.options.scenarioName);
        await assignRuntimeToScenario(
            director,
            this.shoot.compassID,
            this.options.scenarioName
        );
        await delay(60000);
      });

      describe("Actual tests", () => {
        oidcE2ETest();
        commerceMockTest();
      });

      after("Unregister scenario from compass", async function () {
        await unregisterScenarioFromCompass(director, this.options.scenarioName);
      });
      after("Wait for reconciliation", async function () {
        await waitForReconciliation(kcp, this.shoot.name);
      });
    });
  }
  after('Deprovision SKR', async function () {
    try {
      await deprovisionSKR(keb, kcp, this.options.instanceID, deprovisioningTimeout);
      await unregisterRuntimeFromCompass(director, this.options.scenarioName)
    } catch (e) {
      throw new Error(`before hook failed: ${e.toString()}`);
    } finally {
      const runtimeStatus = await kcp.getRuntimeStatusOperations(this.options.instanceID);
      console.log(`\nRuntime status after deprovisioning: ${runtimeStatus}`);
      await kcp.reconcileInformationLog(runtimeStatus);
    }
  });
});