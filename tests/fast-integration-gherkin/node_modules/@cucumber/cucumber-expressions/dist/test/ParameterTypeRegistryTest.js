"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const assert_1 = __importDefault(require("assert"));
const ParameterTypeRegistry_1 = __importDefault(require("../src/ParameterTypeRegistry"));
const ParameterType_1 = __importDefault(require("../src/ParameterType"));
class Name {
    constructor(name) {
        this.name = name;
    }
}
class Person {
    constructor(name) {
        this.name = name;
    }
}
class Place {
    constructor(name) {
        this.name = name;
    }
}
const CAPITALISED_WORD = /[A-Z]+\w+/;
describe('ParameterTypeRegistry', () => {
    let registry;
    beforeEach(() => {
        registry = new ParameterTypeRegistry_1.default();
    });
    it('does not allow more than one preferential parameter type for each regexp', () => {
        registry.defineParameterType(new ParameterType_1.default('name', CAPITALISED_WORD, Name, (s) => new Name(s), true, true));
        registry.defineParameterType(new ParameterType_1.default('person', CAPITALISED_WORD, Person, (s) => new Person(s), true, false));
        try {
            registry.defineParameterType(new ParameterType_1.default('place', CAPITALISED_WORD, Place, (s) => new Place(s), true, true));
            throw new Error('Should have failed');
        }
        catch (err) {
            assert_1.default.strictEqual(err.message, `There can only be one preferential parameter type per regexp. The regexp ${CAPITALISED_WORD} is used for two preferential parameter types, {name} and {place}`);
        }
    });
    it('looks up preferential parameter type by regexp', () => {
        const name = new ParameterType_1.default('name', /[A-Z]+\w+/, null, (s) => new Name(s), true, false);
        const person = new ParameterType_1.default('person', /[A-Z]+\w+/, null, (s) => new Person(s), true, true);
        const place = new ParameterType_1.default('place', /[A-Z]+\w+/, null, (s) => new Place(s), true, false);
        registry.defineParameterType(name);
        registry.defineParameterType(person);
        registry.defineParameterType(place);
        assert_1.default.strictEqual(registry.lookupByRegexp('[A-Z]+\\w+', /([A-Z]+\w+) and ([A-Z]+\w+)/, 'Lisa and Bob'), person);
    });
});
//# sourceMappingURL=ParameterTypeRegistryTest.js.map