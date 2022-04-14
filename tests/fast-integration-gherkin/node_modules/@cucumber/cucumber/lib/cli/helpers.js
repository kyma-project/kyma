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
exports.emitSupportCodeMessages = exports.emitMetaMessage = exports.orderPickleIds = exports.parseGherkinMessageStream = exports.getExpandedArgv = void 0;
const lodash_1 = __importDefault(require("lodash"));
const argv_parser_1 = __importDefault(require("./argv_parser"));
const profile_loader_1 = __importDefault(require("./profile_loader"));
const knuth_shuffle_seeded_1 = __importDefault(require("knuth-shuffle-seeded"));
const value_checker_1 = require("../value_checker");
const option_splitter_1 = __importDefault(require("./option_splitter"));
const messages = __importStar(require("@cucumber/messages"));
const create_meta_1 = __importDefault(require("@cucumber/create-meta"));
const support_code_library_builder_1 = require("../support_code_library_builder");
async function getExpandedArgv({ argv, cwd, }) {
    const { options } = argv_parser_1.default.parse(argv);
    let fullArgv = argv;
    const profileArgv = await new profile_loader_1.default(cwd).getArgv(options.profile);
    if (profileArgv.length > 0) {
        fullArgv = lodash_1.default.concat(argv.slice(0, 2), profileArgv, argv.slice(2));
    }
    return fullArgv;
}
exports.getExpandedArgv = getExpandedArgv;
async function parseGherkinMessageStream({ cwd, eventBroadcaster, eventDataCollector, gherkinMessageStream, order, pickleFilter, }) {
    return await new Promise((resolve, reject) => {
        const result = [];
        gherkinMessageStream.on('data', (envelope) => {
            eventBroadcaster.emit('envelope', envelope);
            if (value_checker_1.doesHaveValue(envelope.pickle)) {
                const pickle = envelope.pickle;
                const pickleId = pickle.id;
                const gherkinDocument = eventDataCollector.getGherkinDocument(pickle.uri);
                if (pickleFilter.matches({ gherkinDocument, pickle })) {
                    result.push(pickleId);
                }
            }
            if (value_checker_1.doesHaveValue(envelope.parseError)) {
                reject(new Error(`Parse error in '${envelope.parseError.source.uri}': ${envelope.parseError.message}`));
            }
        });
        gherkinMessageStream.on('end', () => {
            orderPickleIds(result, order);
            resolve(result);
        });
        gherkinMessageStream.on('error', reject);
    });
}
exports.parseGherkinMessageStream = parseGherkinMessageStream;
// Orders the pickleIds in place - morphs input
function orderPickleIds(pickleIds, order) {
    let [type, seed] = option_splitter_1.default.split(order);
    switch (type) {
        case 'defined':
            break;
        case 'random':
            if (seed === '') {
                seed = Math.floor(Math.random() * 1000 * 1000).toString();
                console.warn(`Random order using seed: ${seed}`);
            }
            knuth_shuffle_seeded_1.default(pickleIds, seed);
            break;
        default:
            throw new Error('Unrecgonized order type. Should be `defined` or `random`');
    }
}
exports.orderPickleIds = orderPickleIds;
async function emitMetaMessage(eventBroadcaster) {
    // eslint-disable-next-line @typescript-eslint/no-var-requires
    const { version } = require('../../package.json');
    eventBroadcaster.emit('envelope', {
        meta: create_meta_1.default('cucumber-js', version, process.env),
    });
}
exports.emitMetaMessage = emitMetaMessage;
function emitParameterTypes(supportCodeLibrary, eventBroadcaster, newId) {
    for (const parameterType of supportCodeLibrary.parameterTypeRegistry
        .parameterTypes) {
        if (support_code_library_builder_1.builtinParameterTypes.includes(parameterType.name)) {
            continue;
        }
        const envelope = {
            parameterType: {
                id: newId(),
                name: parameterType.name,
                preferForRegularExpressionMatch: parameterType.preferForRegexpMatch,
                regularExpressions: parameterType.regexpStrings,
                useForSnippets: parameterType.useForSnippets,
            },
        };
        eventBroadcaster.emit('envelope', envelope);
    }
}
function emitUndefinedParameterTypes(supportCodeLibrary, eventBroadcaster) {
    for (const undefinedParameterType of supportCodeLibrary.undefinedParameterTypes) {
        const envelope = {
            undefinedParameterType,
        };
        eventBroadcaster.emit('envelope', envelope);
    }
}
function emitStepDefinitions(supportCodeLibrary, eventBroadcaster) {
    supportCodeLibrary.stepDefinitions.forEach((stepDefinition) => {
        const envelope = {
            stepDefinition: {
                id: stepDefinition.id,
                pattern: {
                    source: stepDefinition.pattern.toString(),
                    type: typeof stepDefinition.pattern === 'string'
                        ? messages.StepDefinitionPatternType.CUCUMBER_EXPRESSION
                        : messages.StepDefinitionPatternType.REGULAR_EXPRESSION,
                },
                sourceReference: {
                    uri: stepDefinition.uri,
                    location: {
                        line: stepDefinition.line,
                    },
                },
            },
        };
        eventBroadcaster.emit('envelope', envelope);
    });
}
function emitTestCaseHooks(supportCodeLibrary, eventBroadcaster) {
    ;
    []
        .concat(supportCodeLibrary.beforeTestCaseHookDefinitions, supportCodeLibrary.afterTestCaseHookDefinitions)
        .forEach((testCaseHookDefinition) => {
        const envelope = {
            hook: {
                id: testCaseHookDefinition.id,
                tagExpression: testCaseHookDefinition.tagExpression,
                sourceReference: {
                    uri: testCaseHookDefinition.uri,
                    location: {
                        line: testCaseHookDefinition.line,
                    },
                },
            },
        };
        eventBroadcaster.emit('envelope', envelope);
    });
}
function emitTestRunHooks(supportCodeLibrary, eventBroadcaster) {
    ;
    []
        .concat(supportCodeLibrary.beforeTestRunHookDefinitions, supportCodeLibrary.afterTestRunHookDefinitions)
        .forEach((testRunHookDefinition) => {
        const envelope = {
            hook: {
                id: testRunHookDefinition.id,
                sourceReference: {
                    uri: testRunHookDefinition.uri,
                    location: {
                        line: testRunHookDefinition.line,
                    },
                },
            },
        };
        eventBroadcaster.emit('envelope', envelope);
    });
}
function emitSupportCodeMessages({ eventBroadcaster, supportCodeLibrary, newId, }) {
    emitParameterTypes(supportCodeLibrary, eventBroadcaster, newId);
    emitUndefinedParameterTypes(supportCodeLibrary, eventBroadcaster);
    emitStepDefinitions(supportCodeLibrary, eventBroadcaster);
    emitTestCaseHooks(supportCodeLibrary, eventBroadcaster);
    emitTestRunHooks(supportCodeLibrary, eventBroadcaster);
}
exports.emitSupportCodeMessages = emitSupportCodeMessages;
//# sourceMappingURL=helpers.js.map