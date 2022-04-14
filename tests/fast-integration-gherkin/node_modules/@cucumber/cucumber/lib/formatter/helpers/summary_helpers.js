"use strict";
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    Object.defineProperty(o, k2, { enumerable: true, get: function() { return m[k]; } });
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || function (mod) {
    if (mod && mod.__esModule) return mod;
    var result = {};
    if (mod != null) for (var k in mod) if (k !== "default" && Object.prototype.hasOwnProperty.call(mod, k)) __createBinding(result, mod, k);
    __setModuleDefault(result, mod);
    return result;
};
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.formatSummary = void 0;
const lodash_1 = __importDefault(require("lodash"));
const duration_1 = __importDefault(require("duration"));
const messages = __importStar(require("@cucumber/messages"));
const value_checker_1 = require("../../value_checker");
const STATUS_REPORT_ORDER = [
    messages.TestStepResultStatus.FAILED,
    messages.TestStepResultStatus.AMBIGUOUS,
    messages.TestStepResultStatus.UNDEFINED,
    messages.TestStepResultStatus.PENDING,
    messages.TestStepResultStatus.SKIPPED,
    messages.TestStepResultStatus.PASSED,
];
function formatSummary({ colorFns, testCaseAttempts, testRunDuration, }) {
    const testCaseResults = [];
    const testStepResults = [];
    let totalStepDuration = messages.TimeConversion.millisecondsToDuration(0);
    testCaseAttempts.forEach(({ testCase, worstTestStepResult, stepResults }) => {
        Object.values(stepResults).forEach((stepResult) => {
            totalStepDuration = messages.TimeConversion.addDurations(totalStepDuration, stepResult.duration);
        });
        if (!worstTestStepResult.willBeRetried) {
            testCaseResults.push(worstTestStepResult);
            lodash_1.default.each(testCase.testSteps, (testStep) => {
                if (value_checker_1.doesHaveValue(testStep.pickleStepId)) {
                    testStepResults.push(stepResults[testStep.id]);
                }
            });
        }
    });
    const scenarioSummary = getCountSummary({
        colorFns,
        objects: testCaseResults,
        type: 'scenario',
    });
    const stepSummary = getCountSummary({
        colorFns,
        objects: testStepResults,
        type: 'step',
    });
    const durationSummary = `${getDurationSummary(testRunDuration)} (executing steps: ${getDurationSummary(totalStepDuration)})\n`;
    return [scenarioSummary, stepSummary, durationSummary].join('\n');
}
exports.formatSummary = formatSummary;
function getCountSummary({ colorFns, objects, type, }) {
    const counts = lodash_1.default.chain(objects).groupBy('status').mapValues('length').value();
    const total = lodash_1.default.chain(counts).values().sum().value();
    let text = `${total.toString()} ${type}${total === 1 ? '' : 's'}`;
    if (total > 0) {
        const details = [];
        STATUS_REPORT_ORDER.forEach((status) => {
            if (counts[status] > 0) {
                details.push(colorFns.forStatus(status)(`${counts[status].toString()} ${status.toLowerCase()}`));
            }
        });
        text += ` (${details.join(', ')})`;
    }
    return text;
}
function getDurationSummary(durationMsg) {
    const start = new Date(0);
    const end = new Date(messages.TimeConversion.durationToMilliseconds(durationMsg));
    const duration = new duration_1.default(start, end);
    // Use spaces in toString method for readability and to avoid %Ls which is a format
    return duration.toString('%Ms m %S . %L s').replace(/ /g, '');
}
//# sourceMappingURL=summary_helpers.js.map