const {
    ensureCommerceMockWithCompassTestFixture
} = require('../../fast-integration/test/fixtures/commerce-mock');

class CommerceMock {

    constructor() {
        this._initialized = false;
        this._initialPodsRestarts = null;
    }

    static async listAllPodsInCluster() {
        if (!this._initialized){
            this._initialPodsRestarts = await getContainerRestartsForAllNamespaces();
        }

        return this._initialPodsRestarts;
    }

    static async ensureCommerceMockIsSetUp(mockNamespace,
        targetNamespace,
        withCentralApplicationConnectivity = false){
        if(!this._initialized){
            try{
                await this._setupCommerceMock(mockNamespace, targetNamespace, withCentralApplicationConnectivity);
                this._initialized = true;
            }catch(err){
                this._initialized = true;
                console.error(err);
            }
        }
    }

    // TODO: For better readability, have the index.js written as a class here wtih all the methods needed instead of referencing the methods there.
    static async _setupCommerceMock(mockNamespace, 
        targetNamespace, 
        withCentralApplicationConnectivity = false){
            await ensureCommerceMockLocalTestFixture(mockNamespace, targetNamespace, withCentralApplicationConnectivity);
        }
}

module.exports = {
    CommerceMock
}