const {provisionSKR, deprovisionSKR} = require('../../fast-integration/kyma-environment-broker');
const {
    gatherOptions,
    withSuffix,
    withInstanceID
} = require('../../fast-integration/skr-test');
const {KCPWrapper, KCPConfig} = require('../../fast-integration/kcp/client');
const {addScenarioInCompass, assignRuntimeToScenario} = require('../../fast-integration/compass');
const {keb, gardener, director} = require('../../fast-integration/skr-test/provision/provision-skr');
const {
    getEnvOrThrow,
    genRandom,
    initializeK8sClient
} = require('../../fast-integration/utils');
const {BTPOperatorCreds} = require('../../fast-integration/smctl/helpers');
const {unregisterKymaFromCompass} = require('../../fast-integration/compass');
const {getSKRConfig} = require('../../fast-integration/skr-test/helpers');

class SKRSetup {
    constructor() {
        this._skipProvisioning = false;
        this._initialized = false;
        this._skrUpdated = false;
        this._skrAdminsUpdated = false;

        this.updateSkrResponse = null;
        this.updateSkrAdminsResponse = null;
        this.kcp = null;
        this.btpOperatorCreds = null;
        this.shoot = null;
        this.options = null;
    }

    static async provisionSKR() {
        if (!this._initialized){
            this._skipProvisioning = process.env.SKIP_PROVISIONING === 'true';
            if (this._skipProvisioning){
                console.log('Gather information from externally provisioned SKR and prepare the compass resources');
                const instanceID = getEnvOrThrow('INSTANCE_ID');
                let suffix = process.env.TEST_SUFFIX;
                if (suffix === undefined) {
                  suffix = genRandom(4);
                }
                this.options = gatherOptions(
                    withInstanceID(instanceID),
                    withSuffix(suffix),
                );
                this.shoot = await getSKRConfig(instanceID);
                this._initialized = true;
            } else {
                try{
                    const provisioningTimeout = 1000 * 60 * 30; // 30m
                    this.btpOperatorCreds = BTPOperatorCreds.fromEnv();
                    this.kcp = new KCPWrapper(KCPConfig.fromEnv());
                    this.options = gatherOptions();
                    const customParams = {
                      oidc: this.options.oidc0,
                    };
    
                    console.log(this.options.runtimeName);
                    const skr = await provisionSKR(keb, this.kcp, gardener,
                        this.options.instanceID,
                        this.options.runtimeName,
                        null,
                        this.btpOperatorCreds,
                        customParams,
                        provisioningTimeout);
              
                    const runtimeStatus = await this.kcp.getRuntimeStatusOperations(this.options.instanceID);
                    console.log(`\nRuntime status after provisioning: ${runtimeStatus}`);
              
                    this.shoot = skr.shoot;
                    await addScenarioInCompass(director, this.options.scenarioName);
                    await assignRuntimeToScenario(director, this.shoot.compassID, this.options.scenarioName);
                    initializeK8sClient({kubeconfig: this.shoot.kubeconfig});
                    this._initialized = true;
                } catch (e) {
                    throw new Error(`before hook failed: ${e.toString()}`);
                } finally {
                    const runtimeStatus = await this.kcp.getRuntimeStatusOperations(this.options.instanceID);
                    await this.kcp.reconcileInformationLog(runtimeStatus);
                }    
            }
        }
    }

    static async updateSKR(instanceID,
        customParams,
        isMigration = false) {
        if (!this._skrUpdated){
            try{
                this.updateSkrResponse = await keb.updateSKR(instanceID, customParams, null, isMigration);
                this._skrUpdated = true;
            } catch(e) {
                throw new Error(`Failed to update SKR: ${e.toString()}`);
            }
        }
    }

    static async updateSKRAdmins(instanceID,
        customParams,
        isMigration = false) {
        if (!this._skrAdminsUpdated){
            try{
                this.updateSkrAdminsResponse = await keb.updateSKR(instanceID, customParams, null, isMigration);
                this._skrAdminsUpdated = true;
            } catch(e) {
                throw new Error(`Failed to update SKR: ${e.toString()}`);
            }
        }
    }

    static async deprovisionSKR() {
        if (!this._skipProvisioning){
            const deprovisioningTimeout = 1000 * 60 * 95; // 95m

            try {
              await deprovisionSKR(keb, this.kcp, this.options.instanceID, deprovisioningTimeout);
            } catch (e) {
                throw new Error(`before hook failed: ${e.toString()}`);
            } finally {
                const runtimeStatus = await this.kcp.getRuntimeStatusOperations(this.options.instanceID);
                console.log(`\nRuntime status after deprovisioning: ${runtimeStatus}`);
                await this.kcp.reconcileInformationLog(runtimeStatus);
            }
        }
        else{
            console.log('An external SKR cluster was used, de-provisioning skipped');
        }
        
        await unregisterKymaFromCompass(director, this.options.scenarioName);
    }
}

module.exports = {
    SKRSetup
}