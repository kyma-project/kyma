"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const get_color_fns_1 = __importDefault(require("./get_color_fns"));
const javascript_snippet_syntax_1 = __importDefault(require("./step_definition_snippet_builder/javascript_snippet_syntax"));
const json_formatter_1 = __importDefault(require("./json_formatter"));
const message_formatter_1 = __importDefault(require("./message_formatter"));
const path_1 = __importDefault(require("path"));
const progress_bar_formatter_1 = __importDefault(require("./progress_bar_formatter"));
const progress_formatter_1 = __importDefault(require("./progress_formatter"));
const rerun_formatter_1 = __importDefault(require("./rerun_formatter"));
const snippets_formatter_1 = __importDefault(require("./snippets_formatter"));
const step_definition_snippet_builder_1 = __importDefault(require("./step_definition_snippet_builder"));
const summary_formatter_1 = __importDefault(require("./summary_formatter"));
const usage_formatter_1 = __importDefault(require("./usage_formatter"));
const usage_json_formatter_1 = __importDefault(require("./usage_json_formatter"));
const value_checker_1 = require("../value_checker");
const snippet_syntax_1 = require("./step_definition_snippet_builder/snippet_syntax");
const html_formatter_1 = __importDefault(require("./html_formatter"));
const create_require_1 = __importDefault(require("create-require"));
const FormatterBuilder = {
    build(type, options) {
        const FormatterConstructor = FormatterBuilder.getConstructorByType(type, options.cwd);
        const colorFns = get_color_fns_1.default(options.parsedArgvOptions.colorsEnabled);
        const snippetBuilder = FormatterBuilder.getStepDefinitionSnippetBuilder({
            cwd: options.cwd,
            snippetInterface: options.parsedArgvOptions.snippetInterface,
            snippetSyntax: options.parsedArgvOptions.snippetSyntax,
            supportCodeLibrary: options.supportCodeLibrary,
        });
        return new FormatterConstructor(Object.assign({ colorFns,
            snippetBuilder }, options));
    },
    getConstructorByType(type, cwd) {
        switch (type) {
            case 'json':
                return json_formatter_1.default;
            case 'message':
                return message_formatter_1.default;
            case 'html':
                return html_formatter_1.default;
            case 'progress':
                return progress_formatter_1.default;
            case 'progress-bar':
                return progress_bar_formatter_1.default;
            case 'rerun':
                return rerun_formatter_1.default;
            case 'snippets':
                return snippets_formatter_1.default;
            case 'summary':
                return summary_formatter_1.default;
            case 'usage':
                return usage_formatter_1.default;
            case 'usage-json':
                return usage_json_formatter_1.default;
            default:
                return FormatterBuilder.loadCustomFormatter(type, cwd);
        }
    },
    getStepDefinitionSnippetBuilder({ cwd, snippetInterface, snippetSyntax, supportCodeLibrary, }) {
        if (value_checker_1.doesNotHaveValue(snippetInterface)) {
            snippetInterface = snippet_syntax_1.SnippetInterface.Synchronous;
        }
        let Syntax = javascript_snippet_syntax_1.default;
        if (value_checker_1.doesHaveValue(snippetSyntax)) {
            const fullSyntaxPath = path_1.default.resolve(cwd, snippetSyntax);
            Syntax = require(fullSyntaxPath); // eslint-disable-line @typescript-eslint/no-var-requires
        }
        return new step_definition_snippet_builder_1.default({
            snippetSyntax: new Syntax(snippetInterface),
            parameterTypeRegistry: supportCodeLibrary.parameterTypeRegistry,
        });
    },
    loadCustomFormatter(customFormatterPath, cwd) {
        const CustomFormatter = create_require_1.default(cwd)(customFormatterPath);
        if (typeof CustomFormatter === 'function') {
            return CustomFormatter;
        }
        else if (value_checker_1.doesHaveValue(CustomFormatter) &&
            typeof CustomFormatter.default === 'function') {
            return CustomFormatter.default;
        }
        throw new Error(`Custom formatter (${customFormatterPath}) does not export a function`);
    },
};
exports.default = FormatterBuilder;
//# sourceMappingURL=builder.js.map