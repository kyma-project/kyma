"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const assert_1 = __importDefault(require("assert"));
const ParameterType_1 = __importDefault(require("../src/ParameterType"));
const CombinatorialGeneratedExpressionFactory_1 = __importDefault(require("../src/CombinatorialGeneratedExpressionFactory"));
describe('CucumberExpressionGenerator', () => {
    it('generates multiple expressions', () => {
        const parameterTypeCombinations = [
            [
                new ParameterType_1.default('color', /red|blue|yellow/, null, (s) => s, false, true),
                new ParameterType_1.default('csscolor', /red|blue|yellow/, null, (s) => s, false, true),
            ],
            [
                new ParameterType_1.default('date', /\d{4}-\d{2}-\d{2}/, null, (s) => s, false, true),
                new ParameterType_1.default('datetime', /\d{4}-\d{2}-\d{2}/, null, (s) => s, false, true),
                new ParameterType_1.default('timestamp', /\d{4}-\d{2}-\d{2}/, null, (s) => s, false, true),
            ],
        ];
        const factory = new CombinatorialGeneratedExpressionFactory_1.default('I bought a {%s} ball on {%s}', parameterTypeCombinations);
        const expressions = factory.generateExpressions().map((ge) => ge.source);
        assert_1.default.deepStrictEqual(expressions, [
            'I bought a {color} ball on {date}',
            'I bought a {color} ball on {datetime}',
            'I bought a {color} ball on {timestamp}',
            'I bought a {csscolor} ball on {date}',
            'I bought a {csscolor} ball on {datetime}',
            'I bought a {csscolor} ball on {timestamp}',
        ]);
    });
});
//# sourceMappingURL=CombinatorialGeneratedExpressionFactoryTest.js.map