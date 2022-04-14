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
const child_process_1 = require("child_process");
const path_1 = __importDefault(require("path"));
const helpers_1 = require("../helpers");
const messages = __importStar(require("@cucumber/messages"));
const value_checker_1 = require("../../value_checker");
const stopwatch_1 = require("../stopwatch");
const assemble_test_cases_1 = require("../assemble_test_cases");
const runWorkerPath = path_1.default.resolve(__dirname, 'run_worker.js');
class Coordinator {
    constructor({ cwd, eventBroadcaster, eventDataCollector, pickleIds, options, newId, supportCodeLibrary, supportCodePaths, supportCodeRequiredModules, }) {
        this.cwd = cwd;
        this.eventBroadcaster = eventBroadcaster;
        this.eventDataCollector = eventDataCollector;
        this.stopwatch = options.predictableIds
            ? new stopwatch_1.PredictableTestRunStopwatch()
            : new stopwatch_1.RealTestRunStopwatch();
        this.options = options;
        this.newId = newId;
        this.supportCodeLibrary = supportCodeLibrary;
        this.supportCodePaths = supportCodePaths;
        this.supportCodeRequiredModules = supportCodeRequiredModules;
        this.pickleIds = pickleIds;
        this.nextPickleIdIndex = 0;
        this.success = true;
        this.workers = {};
    }
    parseWorkerMessage(worker, message) {
        if (message.ready) {
            this.giveWork(worker);
        }
        else if (value_checker_1.doesHaveValue(message.jsonEnvelope)) {
            const envelope = messages.parseEnvelope(message.jsonEnvelope);
            this.eventBroadcaster.emit('envelope', envelope);
            if (value_checker_1.doesHaveValue(envelope.testCaseFinished)) {
                this.parseTestCaseResult(envelope.testCaseFinished);
            }
        }
        else {
            throw new Error(`Unexpected message from worker: ${JSON.stringify(message)}`);
        }
    }
    startWorker(id, total) {
        const workerProcess = child_process_1.fork(runWorkerPath, [], {
            cwd: this.cwd,
            env: lodash_1.default.assign({}, process.env, {
                CUCUMBER_PARALLEL: 'true',
                CUCUMBER_TOTAL_WORKERS: total,
                CUCUMBER_WORKER_ID: id,
            }),
            stdio: ['inherit', 'inherit', 'inherit', 'ipc'],
        });
        const worker = { closed: false, process: workerProcess };
        this.workers[id] = worker;
        worker.process.on('message', (message) => {
            this.parseWorkerMessage(worker, message);
        });
        worker.process.on('close', (exitCode) => {
            worker.closed = true;
            this.onWorkerProcessClose(exitCode);
        });
        const initializeCommand = {
            initialize: {
                filterStacktraces: this.options.filterStacktraces,
                supportCodePaths: this.supportCodePaths,
                supportCodeRequiredModules: this.supportCodeRequiredModules,
                supportCodeIds: {
                    stepDefinitionIds: this.supportCodeLibrary.stepDefinitions.map((s) => s.id),
                    beforeTestCaseHookDefinitionIds: this.supportCodeLibrary.beforeTestCaseHookDefinitions.map((h) => h.id),
                    afterTestCaseHookDefinitionIds: this.supportCodeLibrary.afterTestCaseHookDefinitions.map((h) => h.id),
                },
                options: this.options,
            },
        };
        worker.process.send(initializeCommand);
    }
    onWorkerProcessClose(exitCode) {
        const success = exitCode === 0;
        if (!success) {
            this.success = false;
        }
        if (lodash_1.default.every(this.workers, 'closed')) {
            const envelope = {
                testRunFinished: {
                    timestamp: this.stopwatch.timestamp(),
                    success,
                },
            };
            this.eventBroadcaster.emit('envelope', envelope);
            this.onFinish(this.success);
        }
    }
    parseTestCaseResult(testCaseFinished) {
        const { worstTestStepResult } = this.eventDataCollector.getTestCaseAttempt(testCaseFinished.testCaseStartedId);
        if (!worstTestStepResult.willBeRetried &&
            this.shouldCauseFailure(worstTestStepResult.status)) {
            this.success = false;
        }
    }
    async run(numberOfWorkers) {
        const envelope = {
            testRunStarted: {
                timestamp: this.stopwatch.timestamp(),
            },
        };
        this.eventBroadcaster.emit('envelope', envelope);
        this.stopwatch.start();
        this.assembledTestCases = await assemble_test_cases_1.assembleTestCases({
            eventBroadcaster: this.eventBroadcaster,
            newId: this.newId,
            pickles: this.pickleIds.map((pickleId) => this.eventDataCollector.getPickle(pickleId)),
            supportCodeLibrary: this.supportCodeLibrary,
        });
        return await new Promise((resolve) => {
            lodash_1.default.times(numberOfWorkers, (id) => this.startWorker(id.toString(), numberOfWorkers));
            this.onFinish = resolve;
        });
    }
    giveWork(worker) {
        if (this.nextPickleIdIndex === this.pickleIds.length) {
            const finalizeCommand = { finalize: true };
            worker.process.send(finalizeCommand);
            return;
        }
        const pickleId = this.pickleIds[this.nextPickleIdIndex];
        this.nextPickleIdIndex += 1;
        const pickle = this.eventDataCollector.getPickle(pickleId);
        const testCase = this.assembledTestCases[pickleId];
        const gherkinDocument = this.eventDataCollector.getGherkinDocument(pickle.uri);
        const retries = helpers_1.retriesForPickle(pickle, this.options);
        const skip = this.options.dryRun || (this.options.failFast && !this.success);
        const runCommand = {
            run: {
                retries,
                skip,
                elapsed: this.stopwatch.duration().nanos(),
                pickle,
                testCase,
                gherkinDocument,
            },
        };
        worker.process.send(runCommand);
    }
    shouldCauseFailure(status) {
        return (lodash_1.default.includes(['AMBIGUOUS', 'FAILED', 'UNDEFINED'], status) ||
            (status === 'PENDING' && this.options.strict));
    }
}
exports.default = Coordinator;
//# sourceMappingURL=coordinator.js.map