const {
    gatherOptions,
    withSuffix,
    withInstanceID,
    provisionSKRInstance
} = require('../../fast-integration/skr-test');
const {KCPWrapper, KCPConfig} = require('../../fast-integration/kcp/client');
const {keb, director} = require('../../fast-integration/skr-test/provision/provision-skr');
const {
    getEnvOrThrow,
    genRandom
} = require('../../fast-integration/utils');
const {BTPOperatorCreds} = require('../../fast-integration/smctl/helpers');
const {unregisterKymaFromCompass} = require('../../fast-integration/compass');
const {getSKRConfig} = require('../../fast-integration/skr-test/helpers');
const {deprovisionSKRInstance} = require('../../fast-integration/skr-test/provision/deprovision-skr');

class SKRSetup {
    constructor() {
        this._skipProvisioning = process.env.SKIP_PROVISIONING === 'true';
        this._initialized = false;
        this._skrUpdated = false;
        this._skrAdminsUpdated = false;
        this._provisioningTimeout = 1000 * 60 * 30; // 30m
        this._deprovisioningTimeout = 1000 * 60 * 95; // 95m

        this.updateSkrResponse = null;
        this.updateSkrAdminsResponse = null;
        this.kcp = null;
        this.btpOperatorCreds = null;
        this.shoot = null;
        this.options = null;
    }

    static async provisionSKR() {
        let globalTimeout = 1000 * 60 * 70; // 70m
        const slowTime = 5000;

        if (!this._skipProvisioning) {
            globalTimeout += this._provisioningTimeout + this._deprovisioningTimeout;
        }
        this.timeout(globalTimeout);
        this.slow(slowTime);

        if (!this._initialized){
            this.options = gatherOptions();
            this.btpOperatorCreds = BTPOperatorCreds.fromEnv();
            this.kcp = new KCPWrapper(KCPConfig.fromEnv());
           
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
            } else {
                this.shoot = await provisionSKRInstance(this.options, this._provisioningTimeout);
            }
            this._initialized = true;
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
            await deprovisionSKRInstance(this.options, this._deprovisioningTimeout);
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