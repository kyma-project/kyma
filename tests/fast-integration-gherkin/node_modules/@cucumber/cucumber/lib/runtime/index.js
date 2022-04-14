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
const lodash_1 = __importStar(require("lodash"));
const helpers_1 = require("../formatter/helpers");
const bluebird_1 = __importDefault(require("bluebird"));
const stack_trace_filter_1 = __importDefault(require("../stack_trace_filter"));
const user_code_runner_1 = __importDefault(require("../user_code_runner"));
const verror_1 = __importDefault(require("verror"));
const helpers_2 = require("./helpers");
const messages = __importStar(require("@cucumber/messages"));
const test_case_runner_1 = __importDefault(require("./test_case_runner"));
const value_checker_1 = require("../value_checker");
const stopwatch_1 = require("./stopwatch");
const assemble_test_cases_1 = require("./assemble_test_cases");
class Runtime {
    constructor({ eventBroadcaster, eventDataCollector, newId, options, pickleIds, supportCodeLibrary, }) {
        this.eventBroadcaster = eventBroadcaster;
        this.eventDataCollector = eventDataCollector;
        this.stopwatch = options.predictableIds
            ? new stopwatch_1.PredictableTestRunStopwatch()
            : new stopwatch_1.RealTestRunStopwatch();
        this.newId = newId;
        this.options = options;
        this.pickleIds = pickleIds;
        this.stackTraceFilter = new stack_trace_filter_1.default();
        this.supportCodeLibrary = supportCodeLibrary;
        this.success = true;
    }
    async runTestRunHooks(definitions, name) {
        if (this.options.dryRun) {
            return;
        }
        await bluebird_1.default.each(definitions, async (hookDefinition) => {
            const { error } = await user_code_runner_1.default.run({
                argsArray: [],
                fn: hookDefinition.code,
                thisArg: null,
                timeoutInMilliseconds: value_checker_1.valueOrDefault(hookDefinition.options.timeout, this.supportCodeLibrary.defaultTimeout),
            });
            if (value_checker_1.doesHaveValue(error)) {
                const location = helpers_1.formatLocation(hookDefinition);
                throw new verror_1.default(error, `${name} hook errored, process exiting: ${location}`);
            }
        });
    }
    async runTestCase(pickleId, testCase) {
        const pickle = this.eventDataCollector.getPickle(pickleId);
        const retries = helpers_2.retriesForPickle(pickle, this.options);
        const skip = this.options.dryRun || (this.options.failFast && !this.success);
        const testCaseRunner = new test_case_runner_1.default({
            eventBroadcaster: this.eventBroadcaster,
            stopwatch: this.stopwatch,
            gherkinDocument: this.eventDataCollector.getGherkinDocument(pickle.uri),
            newId: this.newId,
            pickle,
            testCase,
            retries,
            skip,
            supportCodeLibrary: this.supportCodeLibrary,
            worldParameters: this.options.worldParameters,
        });
        const status = await testCaseRunner.run();
        if (this.shouldCauseFailure(status)) {
            this.success = false;
        }
    }
    async start() {
        if (this.options.filterStacktraces) {
            this.stackTraceFilter.filter();
        }
        const testRunStarted = {
            testRunStarted: {
                timestamp: this.stopwatch.timestamp(),
            },
        };
        this.eventBroadcaster.emit('envelope', testRunStarted);
        this.stopwatch.start();
        await this.runTestRunHooks(this.supportCodeLibrary.beforeTestRunHookDefinitions, 'a BeforeAll');
        const assembledTestCases = await assemble_test_cases_1.assembleTestCases({
            eventBroadcaster: this.eventBroadcaster,
            newId: this.newId,
            pickles: this.pickleIds.map((pickleId) => this.eventDataCollector.getPickle(pickleId)),
            supportCodeLibrary: this.supportCodeLibrary,
        });
        await bluebird_1.default.each(this.pickleIds, async (pickleId) => {
            await this.runTestCase(pickleId, assembledTestCases[pickleId]);
        });
        await this.runTestRunHooks(lodash_1.clone(this.supportCodeLibrary.afterTestRunHookDefinitions).reverse(), 'an AfterAll');
        this.stopwatch.stop();
        const testRunFinished = {
            testRunFinished: {
                timestamp: this.stopwatch.timestamp(),
                success: this.success,
            },
        };
        this.eventBroadcaster.emit('envelope', testRunFinished);
        if (this.options.filterStacktraces) {
            this.stackTraceFilter.unfilter();
        }
        return this.success;
    }
    shouldCauseFailure(status) {
        const failureStatuses = [
            messages.TestStepResultStatus.AMBIGUOUS,
            messages.TestStepResultStatus.FAILED,
            messages.TestStepResultStatus.UNDEFINED,
        ];
        if (this.options.strict)
            failureStatuses.push(messages.TestStepResultStatus.PENDING);
        return lodash_1.default.includes(failureStatuses, status);
    }
}
exports.default = Runtime;
//# sourceMappingURL=index.js.map