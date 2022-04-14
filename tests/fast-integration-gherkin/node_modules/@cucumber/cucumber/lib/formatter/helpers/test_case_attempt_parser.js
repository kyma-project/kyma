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
exports.parseTestCaseAttempt = void 0;
const lodash_1 = __importDefault(require("lodash"));
const keyword_type_1 = require("./keyword_type");
const gherkin_document_parser_1 = require("./gherkin_document_parser");
const pickle_parser_1 = require("./pickle_parser");
const messages = __importStar(require("@cucumber/messages"));
const value_checker_1 = require("../../value_checker");
function parseStep({ isBeforeHook, gherkinStepMap, keyword, keywordType, pickleStep, pickleUri, snippetBuilder, supportCodeLibrary, testStep, testStepResult, testStepAttachments, }) {
    const out = {
        attachments: testStepAttachments,
        keyword: value_checker_1.doesHaveValue(testStep.pickleStepId)
            ? keyword
            : isBeforeHook
                ? 'Before'
                : 'After',
        result: testStepResult,
    };
    if (value_checker_1.doesHaveValue(testStep.hookId)) {
        let hookDefinition;
        if (isBeforeHook) {
            hookDefinition = supportCodeLibrary.beforeTestCaseHookDefinitions.find((x) => x.id === testStep.hookId);
        }
        else {
            hookDefinition = supportCodeLibrary.afterTestCaseHookDefinitions.find((x) => x.id === testStep.hookId);
        }
        out.actionLocation = {
            uri: hookDefinition.uri,
            line: hookDefinition.line,
        };
    }
    if (value_checker_1.doesHaveValue(testStep.stepDefinitionIds) &&
        testStep.stepDefinitionIds.length === 1) {
        const stepDefinition = supportCodeLibrary.stepDefinitions.find((x) => x.id === testStep.stepDefinitionIds[0]);
        out.actionLocation = {
            uri: stepDefinition.uri,
            line: stepDefinition.line,
        };
    }
    if (value_checker_1.doesHaveValue(testStep.pickleStepId)) {
        out.sourceLocation = {
            uri: pickleUri,
            line: gherkinStepMap[pickleStep.astNodeIds[0]].location.line,
        };
        out.text = pickleStep.text;
        if (value_checker_1.doesHaveValue(pickleStep.argument)) {
            out.argument = pickleStep.argument;
        }
    }
    if (testStepResult.status === messages.TestStepResultStatus.UNDEFINED) {
        out.snippet = snippetBuilder.build({ keywordType, pickleStep });
    }
    return out;
}
// Converts a testCaseAttempt into a json object with all data needed for
// displaying it in a pretty format
function parseTestCaseAttempt({ cwd, testCaseAttempt, snippetBuilder, supportCodeLibrary, }) {
    const { testCase, pickle, gherkinDocument } = testCaseAttempt;
    const gherkinStepMap = gherkin_document_parser_1.getGherkinStepMap(gherkinDocument);
    const gherkinScenarioLocationMap = gherkin_document_parser_1.getGherkinScenarioLocationMap(gherkinDocument);
    const pickleStepMap = pickle_parser_1.getPickleStepMap(pickle);
    const relativePickleUri = pickle.uri;
    const parsedTestCase = {
        attempt: testCaseAttempt.attempt,
        name: pickle.name,
        sourceLocation: {
            uri: relativePickleUri,
            line: gherkinScenarioLocationMap[lodash_1.default.last(pickle.astNodeIds)].line,
        },
        worstTestStepResult: testCaseAttempt.worstTestStepResult,
    };
    const parsedTestSteps = [];
    let isBeforeHook = true;
    let previousKeywordType = keyword_type_1.KeywordType.Precondition;
    lodash_1.default.each(testCase.testSteps, (testStep) => {
        const testStepResult = testCaseAttempt.stepResults[testStep.id];
        isBeforeHook = isBeforeHook && value_checker_1.doesHaveValue(testStep.hookId);
        let keyword, keywordType, pickleStep;
        if (value_checker_1.doesHaveValue(testStep.pickleStepId)) {
            pickleStep = pickleStepMap[testStep.pickleStepId];
            keyword = pickle_parser_1.getStepKeyword({ pickleStep, gherkinStepMap });
            keywordType = keyword_type_1.getStepKeywordType({
                keyword,
                language: gherkinDocument.feature.language,
                previousKeywordType,
            });
        }
        const parsedStep = parseStep({
            isBeforeHook,
            gherkinStepMap,
            keyword,
            keywordType,
            pickleStep,
            pickleUri: relativePickleUri,
            snippetBuilder,
            supportCodeLibrary,
            testStep,
            testStepResult,
            testStepAttachments: value_checker_1.valueOrDefault(testCaseAttempt.stepAttachments[testStep.id], []),
        });
        parsedTestSteps.push(parsedStep);
        previousKeywordType = keywordType;
    });
    return {
        testCase: parsedTestCase,
        testSteps: parsedTestSteps,
    };
}
exports.parseTestCaseAttempt = parseTestCaseAttempt;
//# sourceMappingURL=test_case_attempt_parser.js.map