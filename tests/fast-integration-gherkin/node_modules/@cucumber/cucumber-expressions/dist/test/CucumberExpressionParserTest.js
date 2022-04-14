"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const fs_1 = __importDefault(require("fs"));
const js_yaml_1 = __importDefault(require("js-yaml"));
const assert_1 = __importDefault(require("assert"));
const CucumberExpressionParser_1 = __importDefault(require("../src/CucumberExpressionParser"));
const CucumberExpressionError_1 = __importDefault(require("../src/CucumberExpressionError"));
describe('Cucumber expression parser', () => {
    fs_1.default.readdirSync('testdata/ast').forEach((testcase) => {
        const testCaseData = fs_1.default.readFileSync(`testdata/ast/${testcase}`, 'utf-8');
        const expectation = js_yaml_1.default.load(testCaseData);
        it(`${testcase}`, () => {
            const parser = new CucumberExpressionParser_1.default();
            if (expectation.exception == undefined) {
                const node = parser.parse(expectation.expression);
                assert_1.default.deepStrictEqual(JSON.parse(JSON.stringify(node)), // Removes type information.
                JSON.parse(expectation.expected));
            }
            else {
                assert_1.default.throws(() => {
                    parser.parse(expectation.expression);
                }, new CucumberExpressionError_1.default(expectation.exception));
            }
        });
    });
});
//# sourceMappingURL=CucumberExpressionParserTest.js.map