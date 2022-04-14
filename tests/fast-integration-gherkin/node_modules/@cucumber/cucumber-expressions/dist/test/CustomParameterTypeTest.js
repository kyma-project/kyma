'use strict';
var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const assert_1 = __importDefault(require("assert"));
const CucumberExpression_1 = __importDefault(require("../src/CucumberExpression"));
const RegularExpression_1 = __importDefault(require("../src/RegularExpression"));
const ParameterTypeRegistry_1 = __importDefault(require("../src/ParameterTypeRegistry"));
const ParameterType_1 = __importDefault(require("../src/ParameterType"));
class Color {
    /// [color-constructor]
    constructor(name) {
        this.name = name;
    }
}
class CssColor {
    constructor(name) {
        this.name = name;
    }
}
describe('Custom parameter type', () => {
    let parameterTypeRegistry;
    beforeEach(() => {
        parameterTypeRegistry = new ParameterTypeRegistry_1.default();
        /* eslint-disable prettier/prettier */
        /// [add-color-parameter-type]
        parameterTypeRegistry.defineParameterType(new ParameterType_1.default('color', // name
        /red|blue|yellow/, // regexp
        Color, // type
        // type
        s => new Color(s), // transformer
        false, // useForSnippets
        true // preferForRegexpMatch
        ));
        /// [add-color-parameter-type]
        /* eslint-enable prettier/prettier */
    });
    describe('CucumberExpression', () => {
        it('throws exception for illegal character in parameter name', () => {
            assert_1.default.throws(() => new ParameterType_1.default('[string]', /.*/, String, (s) => s, false, true), {
                message: "Illegal character in parameter name {[string]}. Parameter names may not contain '{', '}', '(', ')', '\\' or '/'",
            });
        });
        it('matches parameters with custom parameter type', () => {
            const expression = new CucumberExpression_1.default('I have a {color} ball', parameterTypeRegistry);
            const value = expression.match('I have a red ball')[0].getValue(null);
            assert_1.default.strictEqual(value.name, 'red');
        });
        it('matches parameters with multiple capture groups', () => {
            class Coordinate {
                constructor(x, y, z) {
                    this.x = x;
                    this.y = y;
                    this.z = z;
                }
            }
            parameterTypeRegistry.defineParameterType(new ParameterType_1.default('coordinate', /(\d+),\s*(\d+),\s*(\d+)/, Coordinate, (x, y, z) => new Coordinate(Number(x), Number(y), Number(z)), true, true));
            const expression = new CucumberExpression_1.default('A {int} thick line from {coordinate} to {coordinate}', parameterTypeRegistry);
            const args = expression.match('A 5 thick line from 10,20,30 to 40,50,60');
            const thick = args[0].getValue(null);
            assert_1.default.strictEqual(thick, 5);
            const from = args[1].getValue(null);
            assert_1.default.strictEqual(from.x, 10);
            assert_1.default.strictEqual(from.y, 20);
            assert_1.default.strictEqual(from.z, 30);
            const to = args[2].getValue(null);
            assert_1.default.strictEqual(to.x, 40);
            assert_1.default.strictEqual(to.y, 50);
            assert_1.default.strictEqual(to.z, 60);
        });
        it('matches parameters with custom parameter type using optional capture group', () => {
            parameterTypeRegistry = new ParameterTypeRegistry_1.default();
            parameterTypeRegistry.defineParameterType(new ParameterType_1.default('color', [/red|blue|yellow/, /(?:dark|light) (?:red|blue|yellow)/], Color, (s) => new Color(s), false, true));
            const expression = new CucumberExpression_1.default('I have a {color} ball', parameterTypeRegistry);
            const value = expression.match('I have a dark red ball')[0].getValue(null);
            assert_1.default.strictEqual(value.name, 'dark red');
        });
        it('defers transformation until queried from argument', () => {
            parameterTypeRegistry.defineParameterType(new ParameterType_1.default('throwing', /bad/, null, (s) => {
                throw new Error(`Can't transform [${s}]`);
            }, false, true));
            const expression = new CucumberExpression_1.default('I have a {throwing} parameter', parameterTypeRegistry);
            const args = expression.match('I have a bad parameter');
            assert_1.default.throws(() => args[0].getValue(null), {
                message: "Can't transform [bad]",
            });
        });
        describe('conflicting parameter type', () => {
            it('is detected for type name', () => {
                assert_1.default.throws(() => parameterTypeRegistry.defineParameterType(new ParameterType_1.default('color', /.*/, CssColor, (s) => new CssColor(s), false, true)), { message: 'There is already a parameter type with name color' });
            });
            it('is not detected for type', () => {
                parameterTypeRegistry.defineParameterType(new ParameterType_1.default('whatever', /.*/, Color, (s) => new Color(s), false, false));
            });
            it('is not detected for regexp', () => {
                parameterTypeRegistry.defineParameterType(new ParameterType_1.default('css-color', /red|blue|yellow/, CssColor, (s) => new CssColor(s), true, false));
                assert_1.default.strictEqual(new CucumberExpression_1.default('I have a {css-color} ball', parameterTypeRegistry)
                    .match('I have a blue ball')[0]
                    .getValue(null).constructor, CssColor);
                assert_1.default.strictEqual(new CucumberExpression_1.default('I have a {css-color} ball', parameterTypeRegistry)
                    .match('I have a blue ball')[0]
                    .getValue(null).name, 'blue');
                assert_1.default.strictEqual(new CucumberExpression_1.default('I have a {color} ball', parameterTypeRegistry)
                    .match('I have a blue ball')[0]
                    .getValue(null).constructor, Color);
                assert_1.default.strictEqual(new CucumberExpression_1.default('I have a {color} ball', parameterTypeRegistry)
                    .match('I have a blue ball')[0]
                    .getValue(null).name, 'blue');
            });
        });
        // JavaScript-specific
        it('creates arguments using async transform', () => __awaiter(void 0, void 0, void 0, function* () {
            parameterTypeRegistry = new ParameterTypeRegistry_1.default();
            /// [add-async-parameter-type]
            parameterTypeRegistry.defineParameterType(new ParameterType_1.default('asyncColor', /red|blue|yellow/, Color, (s) => __awaiter(void 0, void 0, void 0, function* () { return new Color(s); }), false, true));
            /// [add-async-parameter-type]
            const expression = new CucumberExpression_1.default('I have a {asyncColor} ball', parameterTypeRegistry);
            const args = yield expression.match('I have a red ball');
            const value = yield args[0].getValue(null);
            assert_1.default.strictEqual(value.name, 'red');
        }));
    });
    describe('RegularExpression', () => {
        it('matches arguments with custom parameter type', () => {
            const expression = new RegularExpression_1.default(/I have a (red|blue|yellow) ball/, parameterTypeRegistry);
            const value = expression.match('I have a red ball')[0].getValue(null);
            assert_1.default.strictEqual(value.constructor, Color);
            assert_1.default.strictEqual(value.name, 'red');
        });
    });
});
//# sourceMappingURL=CustomParameterTypeTest.js.map