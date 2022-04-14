"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.buildParameterType = exports.getDefinitionLineAndUri = void 0;
const lodash_1 = __importDefault(require("lodash"));
const cucumber_expressions_1 = require("@cucumber/cucumber-expressions");
const path_1 = __importDefault(require("path"));
const stacktrace_js_1 = __importDefault(require("stacktrace-js"));
const stack_trace_filter_1 = require("../stack_trace_filter");
const value_checker_1 = require("../value_checker");
function getDefinitionLineAndUri(cwd) {
    let line;
    let uri;
    const stackframes = stacktrace_js_1.default.getSync();
    const stackframe = lodash_1.default.find(stackframes, (frame) => {
        return !stack_trace_filter_1.isFileNameInCucumber(frame.getFileName());
    });
    if (value_checker_1.doesHaveValue(stackframe)) {
        line = stackframe.getLineNumber();
        uri = stackframe.getFileName();
        if (value_checker_1.doesHaveValue(uri)) {
            uri = path_1.default.relative(cwd, uri);
        }
    }
    return {
        line: value_checker_1.valueOrDefault(line, 0),
        uri: value_checker_1.valueOrDefault(uri, 'unknown'),
    };
}
exports.getDefinitionLineAndUri = getDefinitionLineAndUri;
function buildParameterType({ name, regexp, transformer, useForSnippets, preferForRegexpMatch, }) {
    if (typeof useForSnippets !== 'boolean')
        useForSnippets = true;
    if (typeof preferForRegexpMatch !== 'boolean')
        preferForRegexpMatch = false;
    return new cucumber_expressions_1.ParameterType(name, regexp, null, transformer, useForSnippets, preferForRegexpMatch);
}
exports.buildParameterType = buildParameterType;
//# sourceMappingURL=build_helpers.js.map