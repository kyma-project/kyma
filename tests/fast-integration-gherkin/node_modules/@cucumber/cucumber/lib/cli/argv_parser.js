"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const lodash_1 = __importDefault(require("lodash"));
const commander_1 = require("commander");
const path_1 = __importDefault(require("path"));
const gherkin_1 = require("@cucumber/gherkin");
// Using require instead of import so compiled typescript will have the desired folder structure
const { version } = require('../../package.json'); // eslint-disable-line @typescript-eslint/no-var-requires
const ArgvParser = {
    collect(val, memo) {
        memo.push(val);
        return memo;
    },
    mergeJson(option) {
        return function (str, memo) {
            let val;
            try {
                val = JSON.parse(str);
            }
            catch (error) {
                const e = error;
                throw new Error(`${option} passed invalid JSON: ${e.message}: ${str}`);
            }
            if (!lodash_1.default.isPlainObject(val)) {
                throw new Error(`${option} must be passed JSON of an object: ${str}`);
            }
            return lodash_1.default.merge(memo, val);
        };
    },
    mergeTags(value, memo) {
        return memo === '' ? `(${value})` : `${memo} and (${value})`;
    },
    validateCountOption(value, optionName) {
        const numericValue = parseInt(value);
        if (isNaN(numericValue) || numericValue < 0) {
            throw new Error(`${optionName} must be a non negative integer`);
        }
        return numericValue;
    },
    validateLanguage(value) {
        if (!lodash_1.default.includes(lodash_1.default.keys(gherkin_1.dialects), value)) {
            throw new Error(`Unsupported ISO 639-1: ${value}`);
        }
        return value;
    },
    validateRetryOptions(options) {
        if (options.retryTagFilter !== '' && options.retry === 0) {
            throw new Error('a positive --retry count must be specified when setting --retry-tag-filter');
        }
    },
    parse(argv) {
        const program = new commander_1.Command(path_1.default.basename(argv[1]));
        program
            .storeOptionsAsProperties(false)
            .usage('[options] [<GLOB|DIR|FILE[:LINE]>...]')
            .version(version, '-v, --version')
            .option('-b, --backtrace', 'show full backtrace for errors')
            .option('-d, --dry-run', 'invoke formatters without executing steps', false)
            .option('--exit', 'force shutdown of the event loop when the test run has finished: cucumber will call process.exit', false)
            .option('--fail-fast', 'abort the run on first failure', false)
            .option('-f, --format <TYPE[:PATH]>', 'specify the output format, optionally supply PATH to redirect formatter output (repeatable)', ArgvParser.collect, [])
            .option('--format-options <JSON>', 'provide options for formatters (repeatable)', ArgvParser.mergeJson('--format-options'), {})
            .option('--i18n-keywords <ISO 639-1>', 'list language keywords', ArgvParser.validateLanguage, '')
            .option('--i18n-languages', 'list languages', false)
            .option('--language <ISO 639-1>', 'provide the default language for feature files', 'en')
            .option('--name <REGEXP>', 'only execute the scenarios with name matching the expression (repeatable)', ArgvParser.collect, [])
            .option('--no-strict', 'succeed even if there are pending steps')
            .option('--order <TYPE[:SEED]>', 'run scenarios in the specified order. Type should be `defined` or `random`', 'defined')
            .option('-p, --profile <NAME>', 'specify the profile to use (repeatable)', ArgvParser.collect, [])
            .option('--parallel <NUMBER_OF_WORKERS>', 'run in parallel with the given number of workers', (val) => ArgvParser.validateCountOption(val, '--parallel'), 0)
            .option('--predictable-ids', 'Use predictable ids in messages (option ignored if using parallel)', false)
            .option('--publish', 'Publish a report to https://reports.cucumber.io', false)
            .option('--publish-quiet', "Don't print information banner about publishing reports", false)
            .option('-r, --require <GLOB|DIR|FILE>', 'require files before executing features (repeatable)', ArgvParser.collect, [])
            .option('--require-module <NODE_MODULE>', 'require node modules before requiring files (repeatable)', ArgvParser.collect, [])
            .option('--retry <NUMBER_OF_RETRIES>', 'specify the number of times to retry failing test cases (default: 0)', (val) => ArgvParser.validateCountOption(val, '--retry'), 0)
            .option('--retryTagFilter, --retry-tag-filter <EXPRESSION>', `only retries the features or scenarios with tags matching the expression (repeatable).
        This option requires '--retry' to be specified.`, ArgvParser.mergeTags, '')
            .option('-t, --tags <EXPRESSION>', 'only execute the features or scenarios with tags matching the expression (repeatable)', ArgvParser.mergeTags, '')
            .option('--world-parameters <JSON>', 'provide parameters that will be passed to the world constructor (repeatable)', ArgvParser.mergeJson('--world-parameters'), {});
        program.on('--help', () => {
            /* eslint-disable no-console */
            console.log('  For more details please visit https://github.com/cucumber/cucumber-js/blob/master/docs/cli.md\n');
            /* eslint-enable no-console */
        });
        program.parse(argv);
        const options = program.opts();
        ArgvParser.validateRetryOptions(options);
        return {
            options,
            args: program.args,
        };
    },
    lint(fullArgv) {
        if (fullArgv.includes('--retryTagFilter')) {
            console.warn('the argument --retryTagFilter is deprecated and will be removed in a future release; please use --retry-tag-filter');
        }
    },
};
exports.default = ArgvParser;
//# sourceMappingURL=argv_parser.js.map