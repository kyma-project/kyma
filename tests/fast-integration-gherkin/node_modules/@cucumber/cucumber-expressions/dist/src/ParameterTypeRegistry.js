"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const ParameterType_1 = __importDefault(require("./ParameterType"));
const CucumberExpressionGenerator_1 = __importDefault(require("./CucumberExpressionGenerator"));
const Errors_1 = require("./Errors");
const CucumberExpressionError_1 = __importDefault(require("./CucumberExpressionError"));
class ParameterTypeRegistry {
    constructor() {
        this.parameterTypeByName = new Map();
        this.parameterTypesByRegexp = new Map();
        this.defineParameterType(new ParameterType_1.default('int', ParameterTypeRegistry.INTEGER_REGEXPS, Number, (s) => (s === undefined ? null : Number(s)), true, true));
        this.defineParameterType(new ParameterType_1.default('float', ParameterTypeRegistry.FLOAT_REGEXP, Number, (s) => (s === undefined ? null : parseFloat(s)), true, false));
        this.defineParameterType(new ParameterType_1.default('word', ParameterTypeRegistry.WORD_REGEXP, String, (s) => s, false, false));
        this.defineParameterType(new ParameterType_1.default('string', ParameterTypeRegistry.STRING_REGEXP, String, (s1, s2) => (s1 || s2 || '').replace(/\\"/g, '"').replace(/\\'/g, "'"), true, false));
        this.defineParameterType(new ParameterType_1.default('', ParameterTypeRegistry.ANONYMOUS_REGEXP, String, (s) => s, false, true));
    }
    get parameterTypes() {
        return this.parameterTypeByName.values();
    }
    lookupByTypeName(typeName) {
        return this.parameterTypeByName.get(typeName);
    }
    lookupByRegexp(parameterTypeRegexp, expressionRegexp, text) {
        const parameterTypes = this.parameterTypesByRegexp.get(parameterTypeRegexp);
        if (!parameterTypes) {
            return null;
        }
        if (parameterTypes.length > 1 && !parameterTypes[0].preferForRegexpMatch) {
            // We don't do this check on insertion because we only want to restrict
            // ambiguity when we look up by Regexp. Users of CucumberExpression should
            // not be restricted.
            const generatedExpressions = new CucumberExpressionGenerator_1.default(() => this.parameterTypes).generateExpressions(text);
            throw Errors_1.AmbiguousParameterTypeError.forRegExp(parameterTypeRegexp, expressionRegexp, parameterTypes, generatedExpressions);
        }
        return parameterTypes[0];
    }
    defineParameterType(parameterType) {
        if (parameterType.name !== undefined) {
            if (this.parameterTypeByName.has(parameterType.name)) {
                if (parameterType.name.length === 0) {
                    throw new Error(`The anonymous parameter type has already been defined`);
                }
                else {
                    throw new Error(`There is already a parameter type with name ${parameterType.name}`);
                }
            }
            this.parameterTypeByName.set(parameterType.name, parameterType);
        }
        for (const parameterTypeRegexp of parameterType.regexpStrings) {
            if (!this.parameterTypesByRegexp.has(parameterTypeRegexp)) {
                this.parameterTypesByRegexp.set(parameterTypeRegexp, []);
            }
            const parameterTypes = this.parameterTypesByRegexp.get(parameterTypeRegexp);
            const existingParameterType = parameterTypes[0];
            if (parameterTypes.length > 0 &&
                existingParameterType.preferForRegexpMatch &&
                parameterType.preferForRegexpMatch) {
                throw new CucumberExpressionError_1.default('There can only be one preferential parameter type per regexp. ' +
                    `The regexp /${parameterTypeRegexp}/ is used for two preferential parameter types, {${existingParameterType.name}} and {${parameterType.name}}`);
            }
            if (parameterTypes.indexOf(parameterType) === -1) {
                parameterTypes.push(parameterType);
                this.parameterTypesByRegexp.set(parameterTypeRegexp, parameterTypes.sort(ParameterType_1.default.compare));
            }
        }
    }
}
exports.default = ParameterTypeRegistry;
ParameterTypeRegistry.INTEGER_REGEXPS = [/-?\d+/, /\d+/];
ParameterTypeRegistry.FLOAT_REGEXP = /(?=.*\d.*)[-+]?\d*(?:\.(?=\d.*))?\d*(?:\d+[E][+-]?\d+)?/;
ParameterTypeRegistry.WORD_REGEXP = /[^\s]+/;
ParameterTypeRegistry.STRING_REGEXP = /"([^"\\]*(\\.[^"\\]*)*)"|'([^'\\]*(\\.[^'\\]*)*)'/;
ParameterTypeRegistry.ANONYMOUS_REGEXP = /.*/;
module.exports = ParameterTypeRegistry;
//# sourceMappingURL=ParameterTypeRegistry.js.map