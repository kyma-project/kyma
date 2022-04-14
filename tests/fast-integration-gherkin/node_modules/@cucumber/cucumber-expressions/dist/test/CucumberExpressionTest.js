"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const assert_1 = __importDefault(require("assert"));
const CucumberExpression_1 = __importDefault(require("../src/CucumberExpression"));
const ParameterTypeRegistry_1 = __importDefault(require("../src/ParameterTypeRegistry"));
const ParameterType_1 = __importDefault(require("../src/ParameterType"));
const fs_1 = __importDefault(require("fs"));
const js_yaml_1 = __importDefault(require("js-yaml"));
const CucumberExpressionError_1 = __importDefault(require("../src/CucumberExpressionError"));
describe('CucumberExpression', () => {
    fs_1.default.readdirSync('testdata/expression').forEach((testcase) => {
        const testCaseData = fs_1.default.readFileSync(`testdata/expression/${testcase}`, 'utf-8');
        const expectation = js_yaml_1.default.load(testCaseData);
        it(`${testcase}`, () => {
            const parameterTypeRegistry = new ParameterTypeRegistry_1.default();
            if (expectation.exception == undefined) {
                const expression = new CucumberExpression_1.default(expectation.expression, parameterTypeRegistry);
                const matches = expression.match(expectation.text);
                assert_1.default.deepStrictEqual(JSON.parse(JSON.stringify(matches ? matches.map((value) => value.getValue(null)) : null)), // Removes type information.
                JSON.parse(expectation.expected));
            }
            else {
                assert_1.default.throws(() => {
                    const expression = new CucumberExpression_1.default(expectation.expression, parameterTypeRegistry);
                    expression.match(expectation.text);
                }, new CucumberExpressionError_1.default(expectation.exception));
            }
        });
    });
    fs_1.default.readdirSync('testdata/regex').forEach((testcase) => {
        const testCaseData = fs_1.default.readFileSync(`testdata/regex/${testcase}`, 'utf-8');
        const expectation = js_yaml_1.default.load(testCaseData);
        it(`${testcase}`, () => {
            const parameterTypeRegistry = new ParameterTypeRegistry_1.default();
            const expression = new CucumberExpression_1.default(expectation.expression, parameterTypeRegistry);
            assert_1.default.deepStrictEqual(expression.regexp.source, expectation.expected);
        });
    });
    it('documents match arguments', () => {
        const parameterTypeRegistry = new ParameterTypeRegistry_1.default();
        /// [capture-match-arguments]
        const expr = 'I have {int} cuke(s)';
        const expression = new CucumberExpression_1.default(expr, parameterTypeRegistry);
        const args = expression.match('I have 7 cukes');
        assert_1.default.strictEqual(7, args[0].getValue(null));
        /// [capture-match-arguments]
    });
    it('matches float', () => {
        assert_1.default.deepStrictEqual(match('{float}', ''), null);
        assert_1.default.deepStrictEqual(match('{float}', '.'), null);
        assert_1.default.deepStrictEqual(match('{float}', ','), null);
        assert_1.default.deepStrictEqual(match('{float}', '-'), null);
        assert_1.default.deepStrictEqual(match('{float}', 'E'), null);
        assert_1.default.deepStrictEqual(match('{float}', '1,'), null);
        assert_1.default.deepStrictEqual(match('{float}', ',1'), null);
        assert_1.default.deepStrictEqual(match('{float}', '1.'), null);
        assert_1.default.deepStrictEqual(match('{float}', '1'), [1]);
        assert_1.default.deepStrictEqual(match('{float}', '-1'), [-1]);
        assert_1.default.deepStrictEqual(match('{float}', '1.1'), [1.1]);
        assert_1.default.deepStrictEqual(match('{float}', '1,000'), null);
        assert_1.default.deepStrictEqual(match('{float}', '1,000,0'), null);
        assert_1.default.deepStrictEqual(match('{float}', '1,000.1'), null);
        assert_1.default.deepStrictEqual(match('{float}', '1,000,10'), null);
        assert_1.default.deepStrictEqual(match('{float}', '1,0.1'), null);
        assert_1.default.deepStrictEqual(match('{float}', '1,000,000.1'), null);
        assert_1.default.deepStrictEqual(match('{float}', '-1.1'), [-1.1]);
        assert_1.default.deepStrictEqual(match('{float}', '.1'), [0.1]);
        assert_1.default.deepStrictEqual(match('{float}', '-.1'), [-0.1]);
        assert_1.default.deepStrictEqual(match('{float}', '-.10000001'), [-0.10000001]);
        assert_1.default.deepStrictEqual(match('{float}', '1E1'), [1e1]); // precision 1 with scale -1, can not be expressed as a decimal
        assert_1.default.deepStrictEqual(match('{float}', '.1E1'), [1]);
        assert_1.default.deepStrictEqual(match('{float}', 'E1'), null);
        assert_1.default.deepStrictEqual(match('{float}', '-.1E-1'), [-0.01]);
        assert_1.default.deepStrictEqual(match('{float}', '-.1E-2'), [-0.001]);
        assert_1.default.deepStrictEqual(match('{float}', '-.1E+1'), [-1]);
        assert_1.default.deepStrictEqual(match('{float}', '-.1E+2'), [-10]);
        assert_1.default.deepStrictEqual(match('{float}', '-.1E1'), [-1]);
        assert_1.default.deepStrictEqual(match('{float}', '-.10E2'), [-10]);
    });
    it('matches float with zero', () => {
        assert_1.default.deepEqual(match('{float}', '0'), [0]);
    });
    it('exposes source', () => {
        const expr = 'I have {int} cuke(s)';
        assert_1.default.strictEqual(new CucumberExpression_1.default(expr, new ParameterTypeRegistry_1.default()).source, expr);
    });
    it('unmatched optional groups have undefined values', () => {
        const parameterTypeRegistry = new ParameterTypeRegistry_1.default();
        parameterTypeRegistry.defineParameterType(new ParameterType_1.default('textAndOrNumber', /([A-Z]+)?(?: )?([0-9]+)?/, null, function (s1, s2) {
            return [s1, s2];
        }, false, true));
        const expression = new CucumberExpression_1.default('{textAndOrNumber}', parameterTypeRegistry);
        const world = {};
        assert_1.default.deepStrictEqual(expression.match(`TLA`)[0].getValue(world), ['TLA', undefined]);
        assert_1.default.deepStrictEqual(expression.match(`123`)[0].getValue(world), [undefined, '123']);
    });
    // JavaScript-specific
    it('delegates transform to custom object', () => {
        const parameterTypeRegistry = new ParameterTypeRegistry_1.default();
        parameterTypeRegistry.defineParameterType(new ParameterType_1.default('widget', /\w+/, null, function (s) {
            return this.createWidget(s);
        }, false, true));
        const expression = new CucumberExpression_1.default('I have a {widget}', parameterTypeRegistry);
        const world = {
            createWidget(s) {
                return `widget:${s}`;
            },
        };
        const args = expression.match(`I have a bolt`);
        assert_1.default.strictEqual(args[0].getValue(world), 'widget:bolt');
    });
});
const match = (expression, text) => {
    const cucumberExpression = new CucumberExpression_1.default(expression, new ParameterTypeRegistry_1.default());
    const args = cucumberExpression.match(text);
    if (!args) {
        return null;
    }
    return args.map((arg) => arg.getValue(null));
};
//# sourceMappingURL=CucumberExpressionTest.js.map