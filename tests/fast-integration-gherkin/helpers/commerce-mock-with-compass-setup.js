const {
    ensureCommerceMockWithCompassTestFixture,
    deleteMockTestFixture
} = require('../../fast-integration/test/fixtures/commerce-mock');
const{director} = require('../../fast-integration/skr-test/helpers');

class CommerceCompassMock {

    constructor() {
        this._initialized = false;
    }

    static async ensureCommerceWithCompassMockIsSetUp(options){
        if(!this._initialized){
            try{
                await ensureCommerceMockWithCompassTestFixture(director, options.appName, options.scenarioName, 'mocks', options.testNS);
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