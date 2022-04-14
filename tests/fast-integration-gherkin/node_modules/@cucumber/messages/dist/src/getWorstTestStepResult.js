"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.getWorstTestStepResult = void 0;
const messages_1 = require("./messages");
const TimeConversion_1 = require("./TimeConversion");
/**
 * Gets the worst result
 * @param testStepResults
 */
function getWorstTestStepResult(testStepResults) {
    return (testStepResults.slice().sort((r1, r2) => ordinal(r2.status) - ordinal(r1.status))[0] || {
        status: messages_1.TestStepResultStatus.UNKNOWN,
        duration: TimeConversion_1.millisecondsToDuration(0),
        willBeRetried: false,
    });
}
exports.getWorstTestStepResult = getWorstTestStepResult;
function ordinal(status) {
    return [
        messages_1.TestStepResultStatus.UNKNOWN,
        messages_1.TestStepResultStatus.PASSED,
        messages_1.TestStepResultStatus.SKIPPED,
        messages_1.TestStepResultStatus.PENDING,
        messages_1.TestStepResultStatus.UNDEFINED,
        messages_1.TestStepResultStatus.AMBIGUOUS,
        messages_1.TestStepResultStatus.FAILED,
    ].indexOf(status);
}
//# sourceMappingURL=getWorstTestStepResult.js.map