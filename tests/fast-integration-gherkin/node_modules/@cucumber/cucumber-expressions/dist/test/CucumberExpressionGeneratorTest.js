"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const assert_1 = __importDefault(require("assert"));
const CucumberExpressionGenerator_1 = __importDefault(require("../src/CucumberExpressionGenerator"));
const CucumberExpression_1 = __importDefault(require("../src/CucumberExpression"));
const ParameterType_1 = __importDefault(require("../src/ParameterType"));
const ParameterTypeRegistry_1 = __importDefault(require("../src/ParameterTypeRegistry"));
class Currency {
    constructor(s) {
        this.s = s;
    }
}
describe('CucumberExpressionGenerator', () => {
    let parameterTypeRegistry;
    let generator;
    function assertExpression(expectedExpression, expectedArgumentNames, text) {
        const generatedExpression = generator.generateExpressions(text)[0];
        assert_1.default.deepStrictEqual(generatedExpression.parameterNames, expectedArgumentNames);
        assert_1.default.strictEqual(generatedExpression.source, expectedExpression);
        const cucumberExpression = new CucumberExpression_1.default(generatedExpression.source, parameterTypeRegistry);
        const match = cucumberExpression.match(text);
        if (match === null) {
            assert_1.default.fail(`Expected text '${text}' to match generated expression '${generatedExpression.source}'`);
        }
        assert_1.default.strictEqual(match.length, expectedArgumentNames.length);
    }
    beforeEach(() => {
        parameterTypeRegistry = new ParameterTypeRegistry_1.default();
        generator = new CucumberExpressionGenerator_1.default(() => parameterTypeRegistry.parameterTypes);
    });
    it('documents expression generation', () => {
        parameterTypeRegistry = new ParameterTypeRegistry_1.default();
        generator = new CucumberExpressionGenerator_1.default(() => parameterTypeRegistry.parameterTypes);
        const undefinedStepText = 'I have 2 cucumbers and 1.5 tomato';
        const generatedExpression = generator.generateExpressions(undefinedStepText)[0];
        assert_1.default.strictEqual(generatedExpression.source, 'I have {int} cucumbers and {float} tomato');
        assert_1.default.strictEqual(generatedExpression.parameterNames[0], 'int');
        assert_1.default.strictEqual(generatedExpression.parameterTypes[1].name, 'float');
    });
    it('generates expression for no args', () => {
        assertExpression('hello', [], 'hello');
    });
    it('generates expression with escaped left parenthesis', () => {
        assertExpression('\\(iii)', [], '(iii)');
    });
    it('generates expression with escaped left curly brace', () => {
        assertExpression('\\{iii}', [], '{iii}');
    });
    it('generates expression with escaped slashes', () => {
        assertExpression('The {int}\\/{int}\\/{int} hey', ['int', 'int2', 'int3'], 'The 1814/05/17 hey');
    });
    it('generates expression for int float arg', () => {
        assertExpression('I have {int} cukes and {float} euro', ['int', 'float'], 'I have 2 cukes and 1.5 euro');
    });
    it('generates expression for strings', () => {
        assertExpression('I like {string} and {string}', ['string', 'string2'], 'I like "bangers" and \'mash\'');
    });
    it('generates expression with % sign', () => {
        assertExpression('I am {int}%% foobar', ['int'], 'I am 20%% foobar');
    });
    it('generates expression for just int', () => {
        assertExpression('{int}', ['int'], '99999');
    });
    it('numbers only second argument when builtin type is not reserved keyword', () => {
        assertExpression('I have {float} cukes and {float} euro', ['float', 'float2'], 'I have 2.5 cukes and 1.5 euro');
    });
    it('generates expression for custom type', () => {
        parameterTypeRegistry.defineParameterType(new ParameterType_1.default('currency', /[A-Z]{3}/, Currency, (s) => new Currency(s), true, false));
        assertExpression('I have a {currency} account', ['currency'], 'I have a EUR account');
    });
    it('prefers leftmost match when there is overlap', () => {
        parameterTypeRegistry.defineParameterType(new ParameterType_1.default('currency', /c d/, Currency, (s) => new Currency(s), true, false));
        parameterTypeRegistry.defineParameterType(new ParameterType_1.default('date', /b c/, Date, (s) => new Date(s), true, false));
        assertExpression('a {date} d e f g', ['date'], 'a b c d e f g');
    });
    // TODO: prefers widest match
    it('generates all combinations of expressions when several parameter types match', () => {
        parameterTypeRegistry.defineParameterType(new ParameterType_1.default('currency', /x/, null, (s) => new Currency(s), true, false));
        parameterTypeRegistry.defineParameterType(new ParameterType_1.default('date', /x/, null, (s) => new Date(s), true, false));
        const generatedExpressions = generator.generateExpressions('I have x and x and another x');
        const expressions = generatedExpressions.map((e) => e.source);
        assert_1.default.deepStrictEqual(expressions, [
            'I have {currency} and {currency} and another {currency}',
            'I have {currency} and {currency} and another {date}',
            'I have {currency} and {date} and another {currency}',
            'I have {currency} and {date} and another {date}',
            'I have {date} and {currency} and another {currency}',
            'I have {date} and {currency} and another {date}',
            'I have {date} and {date} and another {currency}',
            'I have {date} and {date} and another {date}',
        ]);
    });
    it('exposes parameter type names in generated expression', () => {
        const expression = generator.generateExpressions('I have 2 cukes and 1.5 euro')[0];
        const typeNames = expression.parameterTypes.map((parameter) => parameter.name);
        assert_1.default.deepStrictEqual(typeNames, ['int', 'float']);
    });
    it('matches parameter types with optional capture groups', () => {
        parameterTypeRegistry.defineParameterType(new ParameterType_1.default('optional-flight', /(1st flight)?/, null, (s) => s, true, false));
        parameterTypeRegistry.defineParameterType(new ParameterType_1.default('optional-hotel', /(1 hotel)?/, null, (s) => s, true, false));
        const expression = generator.generateExpressions('I reach Stage 4: 1st flight -1 hotel')[0];
        // While you would expect this to be `I reach Stage {int}: {optional-flight} -{optional-hotel}` the `-1` causes
        // {int} to match just before {optional-hotel}.
        assert_1.default.strictEqual(expression.source, 'I reach Stage {int}: {optional-flight} {int} hotel');
    });
    it('generates at most 256 expressions', () => {
        for (let i = 0; i < 4; i++) {
            parameterTypeRegistry.defineParameterType(new ParameterType_1.default('my-type-' + i, /([a-z] )*?[a-z]/, null, (s) => s, true, false));
        }
        // This would otherwise generate 4^11=419430 expressions and consume just shy of 1.5GB.
        const expressions = generator.generateExpressions('a s i m p l e s t e p');
        assert_1.default.strictEqual(expressions.length, 256);
    });
    it('prefers expression with longest non empty match', () => {
        parameterTypeRegistry.defineParameterType(new ParameterType_1.default('zero-or-more', /[a-z]*/, null, (s) => s, true, false));
        parameterTypeRegistry.defineParameterType(new ParameterType_1.default('exactly-one', /[a-z]/, null, (s) => s, true, false));
        const expressions = generator.generateExpressions('a simple step');
        assert_1.default.strictEqual(expressions.length, 2);
        assert_1.default.strictEqual(expressions[0].source, '{exactly-one} {zero-or-more} {zero-or-more}');
        assert_1.default.strictEqual(expressions[1].source, '{zero-or-more} {zero-or-more} {zero-or-more}');
    });
    it('does not suggest parameter included at the beginning of a word', () => {
        parameterTypeRegistry.defineParameterType(new ParameterType_1.default('direction', /(up|down)/, null, (s) => s, true, false));
        const expressions = generator.generateExpressions('I download a picture');
        assert_1.default.strictEqual(expressions.length, 1);
        assert_1.default.notStrictEqual(expressions[0].source, 'I {direction}load a picture');
        assert_1.default.strictEqual(expressions[0].source, 'I download a picture');
    });
    it('does not suggest parameter included inside a word', () => {
        parameterTypeRegistry.defineParameterType(new ParameterType_1.default('direction', /(up|down)/, null, (s) => s, true, false));
        const expressions = generator.generateExpressions('I watch the muppet show');
        assert_1.default.strictEqual(expressions.length, 1);
        assert_1.default.notStrictEqual(expressions[0].source, 'I watch the m{direction}pet show');
        assert_1.default.strictEqual(expressions[0].source, 'I watch the muppet show');
    });
    it('does not suggest parameter at the end of a word', () => {
        parameterTypeRegistry.defineParameterType(new ParameterType_1.default('direction', /(up|down)/, null, (s) => s, true, false));
        const expressions = generator.generateExpressions('I create a group');
        assert_1.default.strictEqual(expressions.length, 1);
        assert_1.default.notStrictEqual(expressions[0].source, 'I create a gro{direction}');
        assert_1.default.strictEqual(expressions[0].source, 'I create a group');
    });
    it('does suggest parameter that are a full word', () => {
        parameterTypeRegistry.defineParameterType(new ParameterType_1.default('direction', /(up|down)/, null, (s) => s, true, false));
        assert_1.default.strictEqual(generator.generateExpressions('When I go down the road')[0].source, 'When I go {direction} the road');
        assert_1.default.strictEqual(generator.generateExpressions('When I walk up the hill')[0].source, 'When I walk {direction} the hill');
        assert_1.default.strictEqual(generator.generateExpressions('up the hill, the road goes down')[0].source, '{direction} the hill, the road goes {direction}');
    });
    it('does not consider punctuation as being part of a word', () => {
        parameterTypeRegistry.defineParameterType(new ParameterType_1.default('direction', /(up|down)/, null, (s) => s, true, false));
        assert_1.default.strictEqual(generator.generateExpressions('direction is:down')[0].source, 'direction is:{direction}');
        assert_1.default.strictEqual(generator.generateExpressions('direction is down.')[0].source, 'direction is {direction}.');
    });
});
//# sourceMappingURL=CucumberExpressionGeneratorTest.js.map