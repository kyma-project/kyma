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
var __rest = (this && this.__rest) || function (s, e) {
    var t = {};
    for (var p in s) if (Object.prototype.hasOwnProperty.call(s, p) && e.indexOf(p) < 0)
        t[p] = s[p];
    if (s != null && typeof Object.getOwnPropertySymbols === "function")
        for (var i = 0, p = Object.getOwnPropertySymbols(s); i < p.length; i++) {
            if (e.indexOf(p[i]) < 0 && Object.prototype.propertyIsEnumerable.call(s, p[i]))
                t[p[i]] = s[p[i]];
        }
    return t;
};
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.getUsage = void 0;
const lodash_1 = __importDefault(require("lodash"));
const pickle_parser_1 = require("../pickle_parser");
const gherkin_document_parser_1 = require("../gherkin_document_parser");
const messages = __importStar(require("@cucumber/messages"));
const value_checker_1 = require("../../../value_checker");
function buildEmptyMapping(stepDefinitions) {
    const mapping = {};
    stepDefinitions.forEach((stepDefinition) => {
        mapping[stepDefinition.id] = {
            code: stepDefinition.unwrappedCode.toString(),
            line: stepDefinition.line,
            pattern: stepDefinition.expression.source,
            patternType: stepDefinition.expression.constructor.name,
            matches: [],
            uri: stepDefinition.uri,
        };
    });
    return mapping;
}
const unexecutedStatuses = [
    messages.TestStepResultStatus.AMBIGUOUS,
    messages.TestStepResultStatus.SKIPPED,
    messages.TestStepResultStatus.UNDEFINED,
];
function buildMapping({ cwd, stepDefinitions, eventDataCollector, }) {
    const mapping = buildEmptyMapping(stepDefinitions);
    lodash_1.default.each(eventDataCollector.getTestCaseAttempts(), (testCaseAttempt) => {
        const pickleStepMap = pickle_parser_1.getPickleStepMap(testCaseAttempt.pickle);
        const gherkinStepMap = gherkin_document_parser_1.getGherkinStepMap(testCaseAttempt.gherkinDocument);
        testCaseAttempt.testCase.testSteps.forEach((testStep) => {
            if (value_checker_1.doesHaveValue(testStep.pickleStepId) &&
                testStep.stepDefinitionIds.length === 1) {
                const stepDefinitionId = testStep.stepDefinitionIds[0];
                const pickleStep = pickleStepMap[testStep.pickleStepId];
                const gherkinStep = gherkinStepMap[pickleStep.astNodeIds[0]];
                const match = {
                    line: gherkinStep.location.line,
                    text: pickleStep.text,
                    uri: testCaseAttempt.pickle.uri,
                };
                const { duration, status } = testCaseAttempt.stepResults[testStep.id];
                if (!unexecutedStatuses.includes(status) && value_checker_1.doesHaveValue(duration)) {
                    match.duration = duration;
                }
                if (value_checker_1.doesHaveValue(mapping[stepDefinitionId])) {
                    mapping[stepDefinitionId].matches.push(match);
                }
            }
        });
    });
    return mapping;
}
function invertDuration(duration) {
    if (value_checker_1.doesHaveValue(duration)) {
        return -1 * messages.TimeConversion.durationToMilliseconds(duration);
    }
    return 1;
}
function buildResult(mapping) {
    return lodash_1.default.chain(mapping)
        .map((_a) => {
        var { matches } = _a, rest = __rest(_a, ["matches"]);
        const sortedMatches = lodash_1.default.sortBy(matches, [
            (match) => invertDuration(match.duration),
            'text',
        ]);
        const result = Object.assign({ matches: sortedMatches }, rest);
        const durations = lodash_1.default.chain(matches)
            .map((m) => m.duration)
            .compact()
            .value();
        if (durations.length > 0) {
            result.meanDuration = messages.TimeConversion.millisecondsToDuration(lodash_1.default.meanBy(durations, (d) => messages.TimeConversion.durationToMilliseconds(d)));
        }
        return result;
    })
        .sortBy((usage) => invertDuration(usage.meanDuration))
        .value();
}
function getUsage({ cwd, stepDefinitions, eventDataCollector, }) {
    const mapping = buildMapping({ cwd, stepDefinitions, eventDataCollector });
    return buildResult(mapping);
}
exports.getUsage = getUsage;
//# sourceMappingURL=index.js.map