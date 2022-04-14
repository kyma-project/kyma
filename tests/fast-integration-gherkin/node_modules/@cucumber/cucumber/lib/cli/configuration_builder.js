"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const lodash_1 = __importDefault(require("lodash"));
const argv_parser_1 = __importDefault(require("./argv_parser"));
const fs_1 = __importDefault(require("mz/fs"));
const path_1 = __importDefault(require("path"));
const option_splitter_1 = __importDefault(require("./option_splitter"));
const bluebird_1 = __importDefault(require("bluebird"));
const glob_1 = __importDefault(require("glob"));
const util_1 = require("util");
const value_checker_1 = require("../value_checker");
const DEFAULT_CUCUMBER_PUBLISH_URL = 'https://messages.cucumber.io/api/reports';
class ConfigurationBuilder {
    constructor({ argv, cwd }) {
        this.cwd = cwd;
        argv_parser_1.default.lint(argv);
        const parsedArgv = argv_parser_1.default.parse(argv);
        this.args = parsedArgv.args;
        this.options = parsedArgv.options;
    }
    static async build(options) {
        const builder = new ConfigurationBuilder(options);
        return await builder.build();
    }
    async build() {
        const listI18nKeywordsFor = this.options.i18nKeywords;
        const listI18nLanguages = this.options.i18nLanguages;
        const unexpandedFeaturePaths = await this.getUnexpandedFeaturePaths();
        let featurePaths = [];
        let supportCodePaths = [];
        if (listI18nKeywordsFor === '' && !listI18nLanguages) {
            featurePaths = await this.expandFeaturePaths(unexpandedFeaturePaths);
            let unexpandedSupportCodePaths = this.options.require;
            if (unexpandedSupportCodePaths.length === 0) {
                unexpandedSupportCodePaths = this.getFeatureDirectoryPaths(featurePaths);
            }
            supportCodePaths = await this.expandPaths(unexpandedSupportCodePaths, '.js');
        }
        return {
            featureDefaultLanguage: this.options.language,
            featurePaths,
            formats: this.getFormats(),
            formatOptions: this.options.formatOptions,
            publishing: this.isPublishing(),
            listI18nKeywordsFor,
            listI18nLanguages,
            order: this.options.order,
            parallel: this.options.parallel,
            pickleFilterOptions: {
                cwd: this.cwd,
                featurePaths: unexpandedFeaturePaths,
                names: this.options.name,
                tagExpression: this.options.tags,
            },
            predictableIds: this.options.predictableIds,
            profiles: this.options.profile,
            runtimeOptions: {
                dryRun: this.options.dryRun,
                predictableIds: this.options.predictableIds,
                failFast: this.options.failFast,
                filterStacktraces: !this.options.backtrace,
                retry: this.options.retry,
                retryTagFilter: this.options.retryTagFilter,
                strict: this.options.strict,
                worldParameters: this.options.worldParameters,
            },
            shouldExitImmediately: this.options.exit,
            supportCodePaths,
            supportCodeRequiredModules: this.options.requireModule,
            suppressPublishAdvertisement: this.isPublishAdvertisementSuppressed(),
        };
    }
    async expandPaths(unexpandedPaths, defaultExtension) {
        const expandedPaths = await bluebird_1.default.map(unexpandedPaths, async (unexpandedPath) => {
            const matches = await util_1.promisify(glob_1.default)(unexpandedPath, {
                absolute: true,
                cwd: this.cwd,
            });
            const expanded = await bluebird_1.default.map(matches, async (match) => {
                if (path_1.default.extname(match) === '') {
                    return await util_1.promisify(glob_1.default)(`${match}/**/*${defaultExtension}`);
                }
                return [match];
            });
            return lodash_1.default.flatten(expanded);
        });
        return lodash_1.default.flatten(expandedPaths).map((x) => path_1.default.normalize(x));
    }
    async expandFeaturePaths(featurePaths) {
        featurePaths = featurePaths.map((p) => p.replace(/(:\d+)*$/g, '')); // Strip line numbers
        featurePaths = [...new Set(featurePaths)]; // Deduplicate the feature files
        return this.expandPaths(featurePaths, '.feature');
    }
    getFeatureDirectoryPaths(featurePaths) {
        const featureDirs = featurePaths.map((featurePath) => {
            let featureDir = path_1.default.dirname(featurePath);
            let childDir;
            let parentDir = featureDir;
            while (childDir !== parentDir) {
                childDir = parentDir;
                parentDir = path_1.default.dirname(childDir);
                if (path_1.default.basename(parentDir) === 'features') {
                    featureDir = parentDir;
                    break;
                }
            }
            return path_1.default.relative(this.cwd, featureDir);
        });
        return lodash_1.default.uniq(featureDirs);
    }
    isPublishing() {
        return (this.options.publish ||
            this.isTruthyString(process.env.CUCUMBER_PUBLISH_ENABLED) ||
            process.env.CUCUMBER_PUBLISH_TOKEN !== undefined);
    }
    isPublishAdvertisementSuppressed() {
        return (this.options.publishQuiet ||
            this.isTruthyString(process.env.CUCUMBER_PUBLISH_QUIET));
    }
    getFormats() {
        const mapping = { '': 'progress' };
        this.options.format.forEach((format) => {
            const [type, outputTo] = option_splitter_1.default.split(format);
            mapping[outputTo] = type;
        });
        if (this.isPublishing()) {
            const publishUrl = value_checker_1.valueOrDefault(process.env.CUCUMBER_PUBLISH_URL, DEFAULT_CUCUMBER_PUBLISH_URL);
            mapping[publishUrl] = 'message';
        }
        return lodash_1.default.map(mapping, (type, outputTo) => ({ outputTo, type }));
    }
    isTruthyString(s) {
        if (s === undefined) {
            return false;
        }
        return s.match(/^(false|no|0)$/i) === null;
    }
    async getUnexpandedFeaturePaths() {
        if (this.args.length > 0) {
            const nestedFeaturePaths = await bluebird_1.default.map(this.args, async (arg) => {
                const filename = path_1.default.basename(arg);
                if (filename[0] === '@') {
                    const filePath = path_1.default.join(this.cwd, arg);
                    const content = await fs_1.default.readFile(filePath, 'utf8');
                    return lodash_1.default.chain(content).split('\n').map(lodash_1.default.trim).value();
                }
                return [arg];
            });
            const featurePaths = lodash_1.default.flatten(nestedFeaturePaths);
            if (featurePaths.length > 0) {
                return lodash_1.default.compact(featurePaths);
            }
        }
        return ['features/**/*.{feature,feature.md}'];
    }
}
exports.default = ConfigurationBuilder;
//# sourceMappingURL=configuration_builder.js.map