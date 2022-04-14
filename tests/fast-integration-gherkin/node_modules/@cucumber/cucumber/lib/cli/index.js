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
const helpers_1 = require("../formatter/helpers");
const helpers_2 = require("./helpers");
const install_validator_1 = require("./install_validator");
const I18n = __importStar(require("./i18n"));
const configuration_builder_1 = __importDefault(require("./configuration_builder"));
const events_1 = require("events");
const builder_1 = __importDefault(require("../formatter/builder"));
const fs_1 = __importDefault(require("mz/fs"));
const path_1 = __importDefault(require("path"));
const pickle_filter_1 = __importDefault(require("../pickle_filter"));
const bluebird_1 = __importDefault(require("bluebird"));
const coordinator_1 = __importDefault(require("../runtime/parallel/coordinator"));
const runtime_1 = __importDefault(require("../runtime"));
const support_code_library_builder_1 = __importDefault(require("../support_code_library_builder"));
const messages_1 = require("@cucumber/messages");
const value_checker_1 = require("../value_checker");
const gherkin_streams_1 = require("@cucumber/gherkin-streams");
const http_stream_1 = __importDefault(require("../formatter/http_stream"));
const stream_1 = require("stream");
const { incrementing, uuid } = messages_1.IdGenerator;
class Cli {
    constructor({ argv, cwd, stdout, }) {
        this.argv = argv;
        this.cwd = cwd;
        this.stdout = stdout;
    }
    async getConfiguration() {
        const fullArgv = await helpers_2.getExpandedArgv({
            argv: this.argv,
            cwd: this.cwd,
        });
        return await configuration_builder_1.default.build({
            argv: fullArgv,
            cwd: this.cwd,
        });
    }
    async initializeFormatters({ eventBroadcaster, eventDataCollector, formatOptions, formats, supportCodeLibrary, }) {
        const formatters = await bluebird_1.default.map(formats, async ({ type, outputTo }) => {
            let stream = this.stdout;
            if (outputTo !== '') {
                if (outputTo.match(/^https?:\/\//) !== null) {
                    const headers = {};
                    if (process.env.CUCUMBER_PUBLISH_TOKEN !== undefined) {
                        headers.Authorization = `Bearer ${process.env.CUCUMBER_PUBLISH_TOKEN}`;
                    }
                    stream = new http_stream_1.default(outputTo, 'GET', headers);
                    const readerStream = new stream_1.Writable({
                        objectMode: true,
                        write: function (responseBody, encoding, writeCallback) {
                            console.error(responseBody);
                            writeCallback();
                        },
                    });
                    stream.pipe(readerStream);
                }
                else {
                    const fd = await fs_1.default.open(path_1.default.resolve(this.cwd, outputTo), 'w');
                    stream = fs_1.default.createWriteStream(null, { fd });
                }
            }
            stream.on('error', (error) => {
                console.error(error.message);
                process.exit(1);
            });
            const typeOptions = {
                cwd: this.cwd,
                eventBroadcaster,
                eventDataCollector,
                log: stream.write.bind(stream),
                parsedArgvOptions: formatOptions,
                stream,
                cleanup: stream === this.stdout
                    ? async () => await Promise.resolve()
                    : bluebird_1.default.promisify(stream.end.bind(stream)),
                supportCodeLibrary,
            };
            if (value_checker_1.doesNotHaveValue(formatOptions.colorsEnabled)) {
                typeOptions.parsedArgvOptions.colorsEnabled = stream.isTTY;
            }
            if (type === 'progress-bar' && !stream.isTTY) {
                const outputToName = outputTo === '' ? 'stdout' : outputTo;
                console.warn(`Cannot use 'progress-bar' formatter for output to '${outputToName}' as not a TTY. Switching to 'progress' formatter.`);
                type = 'progress';
            }
            return builder_1.default.build(type, typeOptions);
        });
        return async function () {
            await bluebird_1.default.each(formatters, async (formatter) => {
                await formatter.finished();
            });
        };
    }
    getSupportCodeLibrary({ newId, supportCodeRequiredModules, supportCodePaths, }) {
        supportCodeRequiredModules.map((module) => require(module));
        support_code_library_builder_1.default.reset(this.cwd, newId);
        supportCodePaths.forEach((codePath) => {
            try {
                require(codePath);
            }
            catch (e) {
                console.error(e.stack);
                console.error('codepath: ' + codePath);
            }
        });
        return support_code_library_builder_1.default.finalize();
    }
    async run() {
        await install_validator_1.validateInstall(this.cwd);
        const configuration = await this.getConfiguration();
        if (configuration.listI18nLanguages) {
            this.stdout.write(I18n.getLanguages());
            return { shouldExitImmediately: true, success: true };
        }
        if (configuration.listI18nKeywordsFor !== '') {
            this.stdout.write(I18n.getKeywords(configuration.listI18nKeywordsFor));
            return { shouldExitImmediately: true, success: true };
        }
        const newId = configuration.predictableIds && configuration.parallel <= 1
            ? incrementing()
            : uuid();
        const supportCodeLibrary = this.getSupportCodeLibrary({
            newId,
            supportCodePaths: configuration.supportCodePaths,
            supportCodeRequiredModules: configuration.supportCodeRequiredModules,
        });
        const eventBroadcaster = new events_1.EventEmitter();
        const eventDataCollector = new helpers_1.EventDataCollector(eventBroadcaster);
        const cleanup = await this.initializeFormatters({
            eventBroadcaster,
            eventDataCollector,
            formatOptions: configuration.formatOptions,
            formats: configuration.formats,
            supportCodeLibrary,
        });
        await helpers_2.emitMetaMessage(eventBroadcaster);
        const gherkinMessageStream = gherkin_streams_1.GherkinStreams.fromPaths(configuration.featurePaths, {
            defaultDialect: configuration.featureDefaultLanguage,
            newId,
            relativeTo: this.cwd,
        });
        let pickleIds = [];
        if (configuration.featurePaths.length > 0) {
            pickleIds = await helpers_2.parseGherkinMessageStream({
                cwd: this.cwd,
                eventBroadcaster,
                eventDataCollector,
                gherkinMessageStream,
                order: configuration.order,
                pickleFilter: new pickle_filter_1.default(configuration.pickleFilterOptions),
            });
        }
        helpers_2.emitSupportCodeMessages({
            eventBroadcaster,
            supportCodeLibrary,
            newId,
        });
        let success;
        if (configuration.parallel > 1) {
            const parallelRuntimeCoordinator = new coordinator_1.default({
                cwd: this.cwd,
                eventBroadcaster,
                eventDataCollector,
                options: configuration.runtimeOptions,
                newId,
                pickleIds,
                supportCodeLibrary,
                supportCodePaths: configuration.supportCodePaths,
                supportCodeRequiredModules: configuration.supportCodeRequiredModules,
            });
            success = await parallelRuntimeCoordinator.run(configuration.parallel);
        }
        else {
            const runtime = new runtime_1.default({
                eventBroadcaster,
                eventDataCollector,
                options: configuration.runtimeOptions,
                newId,
                pickleIds,
                supportCodeLibrary,
            });
            success = await runtime.start();
        }
        await cleanup();
        return {
            shouldExitImmediately: configuration.shouldExitImmediately,
            success,
        };
    }
}
exports.default = Cli;
//# sourceMappingURL=index.js.map