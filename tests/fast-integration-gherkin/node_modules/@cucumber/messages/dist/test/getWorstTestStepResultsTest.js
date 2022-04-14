"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const getWorstTestStepResult_1 = require("../src/getWorstTestStepResult");
const messages_1 = require("../src/messages");
const assert_1 = __importDefault(require("assert"));
describe('getWorstTestStepResult', () => {
    it('returns a FAILED result for PASSED,FAILED,PASSED', () => {
        const result = getWorstTestStepResult_1.getWorstTestStepResult([
            {
                status: messages_1.TestStepResultStatus.PASSED,
                duration: { seconds: 0, nanos: 0 },
                willBeRetried: false,
            },
            {
                status: messages_1.TestStepResultStatus.FAILED,
                duration: { seconds: 0, nanos: 0 },
                willBeRetried: false,
            },
            {
                status: messages_1.TestStepResultStatus.PASSED,
                duration: { seconds: 0, nanos: 0 },
                willBeRetried: false,
            },
        ]);
        assert_1.default.strictEqual(result.status, messages_1.TestStepResultStatus.FAILED);
    });
});
//# sourceMappingURL=getWorstTestStepResultsTest.js.map