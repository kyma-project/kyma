"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const fs_1 = __importDefault(require("fs"));
const js_yaml_1 = __importDefault(require("js-yaml"));
const CucumberExpressionTokenizer_1 = __importDefault(require("../src/CucumberExpressionTokenizer"));
const assert_1 = __importDefault(require("assert"));
const CucumberExpressionError_1 = __importDefault(require("../src/CucumberExpressionError"));
describe('Cucumber expression tokenizer', () => {
    fs_1.default.readdirSync('testdata/tokens').forEach((testcase) => {
        const testCaseData = fs_1.default.readFileSync(`testdata/tokens/${testcase}`, 'utf-8');
        const expectation = js_yaml_1.default.load(testCaseData);
        it(`${testcase}`, () => {
            const tokenizer = new CucumberExpressionTokenizer_1.default();
            if (expectation.exception == undefined) {
                const tokens = tokenizer.tokenize(expectation.expression);
                assert_1.default.deepStrictEqual(JSON.parse(JSON.stringify(tokens)), // Removes type information.
                JSON.parse(expectation.expected));
            }
            else {
                assert_1.default.throws(() => {
                    tokenizer.tokenize(expectation.expression);
                }, new CucumberExpressionError_1.default(expectation.exception));
            }
        });
    });
});
//# sourceMappingURL=CucumberExpressionTokenizerTest.js.map