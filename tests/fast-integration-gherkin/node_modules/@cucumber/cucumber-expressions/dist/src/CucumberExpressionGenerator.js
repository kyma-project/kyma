"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const ParameterTypeMatcher_1 = __importDefault(require("./ParameterTypeMatcher"));
const ParameterType_1 = __importDefault(require("./ParameterType"));
const util_1 = __importDefault(require("util"));
const CombinatorialGeneratedExpressionFactory_1 = __importDefault(require("./CombinatorialGeneratedExpressionFactory"));
class CucumberExpressionGenerator {
    constructor(parameterTypes) {
        this.parameterTypes = parameterTypes;
    }
    generateExpressions(text) {
        const parameterTypeCombinations = [];
        const parameterTypeMatchers = this.createParameterTypeMatchers(text);
        let expressionTemplate = '';
        let pos = 0;
        // eslint-disable-next-line no-constant-condition
        while (true) {
            let matchingParameterTypeMatchers = [];
            for (const parameterTypeMatcher of parameterTypeMatchers) {
                const advancedParameterTypeMatcher = parameterTypeMatcher.advanceTo(pos);
                if (advancedParameterTypeMatcher.find) {
                    matchingParameterTypeMatchers.push(advancedParameterTypeMatcher);
                }
            }
            if (matchingParameterTypeMatchers.length > 0) {
                matchingParameterTypeMatchers = matchingParameterTypeMatchers.sort(ParameterTypeMatcher_1.default.compare);
                // Find all the best parameter type matchers, they are all candidates.
                const bestParameterTypeMatcher = matchingParameterTypeMatchers[0];
                const bestParameterTypeMatchers = matchingParameterTypeMatchers.filter((m) => ParameterTypeMatcher_1.default.compare(m, bestParameterTypeMatcher) === 0);
                // Build a list of parameter types without duplicates. The reason there
                // might be duplicates is that some parameter types have more than one regexp,
                // which means multiple ParameterTypeMatcher objects will have a reference to the
                // same ParameterType.
                // We're sorting the list so preferential parameter types are listed first.
                // Users are most likely to want these, so they should be listed at the top.
                let parameterTypes = [];
                for (const parameterTypeMatcher of bestParameterTypeMatchers) {
                    if (parameterTypes.indexOf(parameterTypeMatcher.parameterType) === -1) {
                        parameterTypes.push(parameterTypeMatcher.parameterType);
                    }
                }
                parameterTypes = parameterTypes.sort(ParameterType_1.default.compare);
                parameterTypeCombinations.push(parameterTypes);
                expressionTemplate += escape(text.slice(pos, bestParameterTypeMatcher.start));
                expressionTemplate += '{%s}';
                pos = bestParameterTypeMatcher.start + bestParameterTypeMatcher.group.length;
            }
            else {
                break;
            }
            if (pos >= text.length) {
                break;
            }
        }
        expressionTemplate += escape(text.slice(pos));
        return new CombinatorialGeneratedExpressionFactory_1.default(expressionTemplate, parameterTypeCombinations).generateExpressions();
    }
    /**
     * @deprecated
     */
    generateExpression(text) {
        return util_1.default.deprecate(() => this.generateExpressions(text)[0], 'CucumberExpressionGenerator.generateExpression: Use CucumberExpressionGenerator.generateExpressions instead')();
    }
    createParameterTypeMatchers(text) {
        let parameterMatchers = [];
        for (const parameterType of this.parameterTypes()) {
            if (parameterType.useForSnippets) {
                parameterMatchers = parameterMatchers.concat(CucumberExpressionGenerator.createParameterTypeMatchers2(parameterType, text));
            }
        }
        return parameterMatchers;
    }
    static createParameterTypeMatchers2(parameterType, text) {
        // TODO: [].map
        const result = [];
        for (const regexp of parameterType.regexpStrings) {
            result.push(new ParameterTypeMatcher_1.default(parameterType, regexp, text));
        }
        return result;
    }
}
exports.default = CucumberExpressionGenerator;
function escape(s) {
    return s
        .replace(/%/g, '%%') // for util.format
        .replace(/\(/g, '\\(')
        .replace(/{/g, '\\{')
        .replace(/\//g, '\\/');
}
module.exports = CucumberExpressionGenerator;
//# sourceMappingURL=CucumberExpressionGenerator.js.map