const {
    ensureCommerceMockWithCompassTestFixture,
    deleteMockTestFixture
} = require('../../fast-integration/test/fixtures/commerce-mock');

class CommerceCompassMock {

    constructor() {
        this._initialized = false;

        this._director = new DirectorClient(DirectorConfig.fromEnv());
    }

    static async ensureCommerceWithCompassMockIsSetUp(options){
        if(!this._initialized){
            try{
                await ensureCommerceMockWithCompassTestFixture(this._director, options.appName, options.scenarioName, 'mocks', options.testNS);
                this._initialized = true;
            }catch(err){
                console.error(err);
            }
        }
    }

    static async deleteCommerceMockResources(testNamespace){
        await deleteMockTestFixture('mocks', testNamespace);
    }
}

module.exports = {
    CommerceCompassMock
}