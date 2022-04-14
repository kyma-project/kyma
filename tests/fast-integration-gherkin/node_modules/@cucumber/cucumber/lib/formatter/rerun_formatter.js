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
const lodash_1 = __importDefault(require("lodash"));
const _1 = __importDefault(require("./"));
const gherkin_document_parser_1 = require("./helpers/gherkin_document_parser");
const value_checker_1 = require("../value_checker");
const messages = __importStar(require("@cucumber/messages"));
const DEFAULT_SEPARATOR = '\n';
class RerunFormatter extends _1.default {
    constructor(options) {
        super(options);
        options.eventBroadcaster.on('envelope', (envelope) => {
            if (value_checker_1.doesHaveValue(envelope.testRunFinished)) {
                this.logFailedTestCases();
            }
        });
        const rerunOptions = value_checker_1.valueOrDefault(options.parsedArgvOptions.rerun, {});
        this.separator = value_checker_1.valueOrDefault(rerunOptions.separator, DEFAULT_SEPARATOR);
    }
    logFailedTestCases() {
        const mapping = {};
        lodash_1.default.each(this.eventDataCollector.getTestCaseAttempts(), ({ gherkinDocument, pickle, worstTestStepResult }) => {
            if (worstTestStepResult.status !== messages.TestStepResultStatus.PASSED) {
                const relativeUri = pickle.uri;
                const line = gherkin_document_parser_1.getGherkinScenarioLocationMap(gherkinDocument)[lodash_1.default.last(pickle.astNodeIds)].line;
                if (value_checker_1.doesNotHaveValue(mapping[relativeUri])) {
                    mapping[relativeUri] = [];
                }
                mapping[relativeUri].push(line);
            }
        });
        const text = lodash_1.default.chain(mapping)
            .map((lines, uri) => `${uri}:${lines.join(':')}`)
            .join(this.separator)
            .value();
        this.log(text);
    }
}
exports.default = RerunFormatter;
//# sourceMappingURL=rerun_formatter.js.map