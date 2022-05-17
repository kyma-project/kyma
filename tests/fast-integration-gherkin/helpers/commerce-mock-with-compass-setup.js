const {
    ensureCommerceMockWithCompassTestFixture,
    deleteMockTestFixture
} = require('../../fast-integration/test/fixtures/commerce-mock');
const {DirectorClient, DirectorConfig} = require('../../fast-integration/compass');

class CommerceCompassMock {

    constructor() {
        this._initialized = false;
    }

    static async ensureCommerceWithCompassMockIsSetUp(options){
        if(!this._initialized){
            try{
                this._director = new DirectorClient(DirectorConfig.fromEnv());
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