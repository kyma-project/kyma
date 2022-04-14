"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const fs_1 = __importDefault(require("fs"));
const assert_1 = __importDefault(require("assert"));
const CucumberExpression_1 = __importDefault(require("../src/CucumberExpression"));
const RegularExpression_1 = __importDefault(require("../src/RegularExpression"));
const ParameterTypeRegistry_1 = __importDefault(require("../src/ParameterTypeRegistry"));
describe('examples.txt', () => {
    const match = (expressionText, text) => {
        const m = /^\/(.*)\/$/.exec(expressionText);
        const expression = m
            ? new RegularExpression_1.default(new RegExp(m[1]), new ParameterTypeRegistry_1.default())
            : new CucumberExpression_1.default(expressionText, new ParameterTypeRegistry_1.default());
        const args = expression.match(text);
        if (!args) {
            return null;
        }
        return args.map((arg) => arg.getValue(null));
    };
    const examples = fs_1.default.readFileSync('examples.txt', 'utf-8');
    const chunks = examples.split(/^---/m);
    for (const chunk of chunks) {
        const [expressionText, text, expectedArgs] = chunk.trim().split(/\n/m);
        it(`Works with: ${expressionText}`, () => {
            assert_1.default.deepStrictEqual(JSON.stringify(match(expressionText, text)), expectedArgs);
        });
    }
});
//# sourceMappingURL=ExpressionExamplesTest.js.map