const {getEnvOrThrow} = require("../utils");
const {gatherOptions, withInstanceID, withOIDC0} = require('./');
const {provisionSKR, deprovisionSKR} = require('../kyma-environment-broker');
const {keb, gardener, director} = require('./helpers');
const {initializeK8sClient} = require('../utils');
const {unregisterKymaFromCompass, addScenarioInCompass, assignRuntimeToScenario} = require('../compass');
const {oidcE2ETest, commerceMockTest} = require('./skr-test-N');
const {KCPWrapper, KCPConfig} = require('../kcp/client');

const SKR_CLUSTER = process.env.SKR_CLUSTER === "true";
let N = 3;// process.env.N;

const kcp = new KCPWrapper(KCPConfig.fromEnv());

const delay = millis => new Promise((resolve, reject) => {
  setTimeout(_ => resolve(), millis)
});

describe('Execute SKR test',  async function () {
  this.timeout(60 * 60 * 1000 * 3); // 3h
  this.slow(5000);

  const provisioningTimeout = 1000 * 60 * 30; // 30m
  const deprovisioningTimeout = 1000 * 60 * 95; // 95m

  before('Provision SKR', async function () {
    try {
      if (!SKR_CLUSTER) {
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
      } else {
        this.shoot = await gardener.getShoot(getEnvOrThrow("SHOOT_NAME"));

        this.options = gatherOptions(
            withInstanceID(getEnvOrThrow("INSTANCE_ID")),
            withOIDC0(this.shoot.oidcConfig))
      }
      const runtimeStatus = await kcp.getRuntimeStatusOperations(this.options.instanceID);
      console.log(`\nRuntime status after provisioning: ${runtimeStatus}`);
      // await addScenarioInCompass(director, this.options.scenarioName);
      // await assignRuntimeToScenario(director, this.shoot.compassID, this.options.scenarioName);
      // initializeK8sClient({kubeconfig: this.shoot.kubeconfig});
///
      this.lastReconciliation = await kcp.getLastReconciliation(this.shoot.name)
///
    } catch (e) {
      throw new Error(`before hook failed: ${e.toString()}`);
    } finally {
      //const runtimeStatus = await kcp.getRuntimeStatusOperations(this.options.instanceID);
      //await kcp.reconcileInformationLog(runtimeStatus);
    }
  });

  for(let i=0; i<N; i++) {
    describe(`Loop ${i+1}`, function () {
      before('Before', async function () {
        this.options = gatherOptions(
            withInstanceID(this.options.instanceID),
            withOIDC0(this.shoot.oidcConfig));
        await addScenarioInCompass(director, this.options.scenarioName);
        await assignRuntimeToScenario(director, this.shoot.compassID, this.options.scenarioName);
        initializeK8sClient({kubeconfig: this.shoot.kubeconfig});
        await delay(60000);
      })
      oidcE2ETest();
      after('After', async function () {
        await unregisterKymaFromCompass(director, this.options.scenarioName);
      })
    });
  }

  // while (N > 0) {
  // for(let i=0; i<N; i++) {
  //   const myPromise = new Promise((resolve, reject) => {
  //     describe(`Loop ${N}`, async function () {
  //       it(`Assure initial OIDC config is applied on shoot cluster ${N} `, async function () {
  //         console.log("it")
  //         let l = await kcp.getLastReconciliation(this.shoot.name)
  //         console.log(l.schedulingID, l.status)
  //
  //         while (!(this.lastReconciliation.schedulingID !== l.schedulingID && l.status === "ready")) {
  //           console.log(`Last Reconciliation is  ${this.lastReconciliation.schedulingID}`)
  //           console.log(`Current Reconciliation is  ${l.schedulingID}`)
  //           console.log(`Current Reconciliation status is  ${l.status}`)
  //           console.log(`condition  is  ${!(this.lastReconciliation.schedulingID !== l.schedulingID && l.status === "ready")}`)
  //           await delay(5000)
  //           l = await kcp.getLastReconciliation(this.shoot.name)
  //         }
  //
  //         this.lastReconciliation = l;
  //         oidcE2ETest();
  //         //N--;
  //         resolve();
  //         console.log("it -  I am done")
  //
  //       });
  //     });
  //   });
  //   await myPromise;
  // }

  after('Deprovision SKR', async function () {
    try {
      if (!SKR_CLUSTER) {
        await deprovisionSKR(keb, kcp, this.options.instanceID, deprovisioningTimeout);
      }
    } catch (e) {
      throw new Error(`before hook failed: ${e.toString()}`);
    } finally {
      // const runtimeStatus = await kcp.getRuntimeStatusOperations(this.options.instanceID);
      // console.log(`\nRuntime status after deprovisioning: ${runtimeStatus}`);
      //await kcp.reconcileInformationLog(runtimeStatus);
      //await unregisterKymaFromCompass(director, this.options.scenarioName);
    }
  });
});


// let l = await kcp.getLastReconciliation(this.shoot.name)
// while (!(this.lastReconciliation.schedulingID !== l.schedulingID && l.status === "ready")) {
//   console.log(`Lasts Reeconciliation is  ${this.lastReconciliation.schedulingID}`)
//   console.log(`Current Reeconciliation is  ${l.schedulingID}`)
//   console.log(`Current Reeconciliation status is  ${l.status}`)
//   console.log(`condition  is  ${!(this.lastReconciliation.schedulingID !== l.schedulingID && l.status === "ready")}`)
//   await delay(5000)
//   l = await kcp.getLastReconciliation(this.shoot.name)
// }
// this.lastReconciliation = l;
