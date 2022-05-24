import { IOptions } from "../Interfaces/IOptions";

const {
    ensureCommerceMockWithCompassTestFixture,
    deleteMockTestFixture
} = require('../../../fast-integration/test/fixtures/commerce-mock');
const {DirectorClient, DirectorConfig} = require('../../../fast-integration/compass');

class CommerceCompassMock {
    private static _initialized = false;
    private static _director: any;

    static async ensureCommerceWithCompassMockIsSetUp(options: IOptions){
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

    static async deleteCommerceMockResources(testNamespace: string){
        await deleteMockTestFixture('mocks', testNamespace);
    }
}

export {
    CommerceCompassMock
}