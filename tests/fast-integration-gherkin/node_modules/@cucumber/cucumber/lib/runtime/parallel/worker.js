"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const helpers_1 = require("../../formatter/helpers");
const events_1 = require("events");
const bluebird_1 = __importDefault(require("bluebird"));
const stack_trace_filter_1 = __importDefault(require("../../stack_trace_filter"));
const support_code_library_builder_1 = __importDefault(require("../../support_code_library_builder"));
const test_case_runner_1 = __importDefault(require("../test_case_runner"));
const user_code_runner_1 = __importDefault(require("../../user_code_runner"));
const messages_1 = require("@cucumber/messages");
const value_checker_1 = require("../../value_checker");
const stopwatch_1 = require("../stopwatch");
const durations_1 = require("durations");
const { uuid } = messages_1.IdGenerator;
class Worker {
    constructor({ cwd, exit, id, sendMessage, }) {
        this.id = id;
        this.newId = uuid();
        this.cwd = cwd;
        this.exit = exit;
        this.sendMessage = sendMessage;
        this.eventBroadcaster = new events_1.EventEmitter();
        this.stackTraceFilter = new stack_trace_filter_1.default();
        this.eventBroadcaster.on('envelope', (envelope) => {
            this.sendMessage({
                jsonEnvelope: JSON.stringify(envelope),
            });
        });
    }
    async initialize({ filterStacktraces, supportCodeRequiredModules, supportCodePaths, supportCodeIds, options, }) {
        supportCodeRequiredModules.map((module) => require(module));
        support_code_library_builder_1.default.reset(this.cwd, this.newId);
        supportCodePaths.forEach((codePath) => require(codePath));
        this.supportCodeLibrary = support_code_library_builder_1.default.finalize(supportCodeIds);
        this.worldParameters = options.worldParameters;
        this.options = options;
        this.filterStacktraces = filterStacktraces;
        if (this.filterStacktraces) {
            this.stackTraceFilter.filter();
        }
        await this.runTestRunHooks(this.supportCodeLibrary.beforeTestRunHookDefinitions, 'a BeforeAll');
        this.sendMessage({ ready: true });
    }
    async finalize() {
        await this.runTestRunHooks(this.supportCodeLibrary.afterTestRunHookDefinitions, 'an AfterAll');
        if (this.filterStacktraces) {
            this.stackTraceFilter.unfilter();
        }
        this.exit(0);
    }
    async receiveMessage(message) {
        if (value_checker_1.doesHaveValue(message.initialize)) {
            await this.initialize(message.initialize);
        }
        else if (message.finalize) {
            await this.finalize();
        }
        else if (value_checker_1.doesHaveValue(message.run)) {
            await this.runTestCase(message.run);
        }
    }
    async runTestCase({ gherkinDocument, pickle, testCase, elapsed, retries, skip, }) {
        const stopwatch = this.options.predictableIds
            ? new stopwatch_1.PredictableTestRunStopwatch()
            : new stopwatch_1.RealTestRunStopwatch();
        stopwatch.from(durations_1.duration(elapsed));
        const testCaseRunner = new test_case_runner_1.default({
            eventBroadcaster: this.eventBroadcaster,
            stopwatch,
            gherkinDocument,
            newId: this.newId,
            pickle,
            testCase,
            retries,
            skip,
            supportCodeLibrary: this.supportCodeLibrary,
            worldParameters: this.worldParameters,
        });
        await testCaseRunner.run();
        this.sendMessage({ ready: true });
    }
    async runTestRunHooks(testRunHookDefinitions, name) {
        if (this.options.dryRun) {
            return;
        }
        await bluebird_1.default.each(testRunHookDefinitions, async (hookDefinition) => {
            const { error } = await user_code_runner_1.default.run({
                argsArray: [],
                fn: hookDefinition.code,
                thisArg: null,
                timeoutInMilliseconds: value_checker_1.valueOrDefault(hookDefinition.options.timeout, this.supportCodeLibrary.defaultTimeout),
            });
            if (value_checker_1.doesHaveValue(error)) {
                const location = helpers_1.formatLocation(hookDefinition);
                this.exit(1, error, `${name} hook errored on worker ${this.id}, process exiting: ${location}`);
            }
        });
    }
}
exports.default = Worker;
//# sourceMappingURL=worker.js.map