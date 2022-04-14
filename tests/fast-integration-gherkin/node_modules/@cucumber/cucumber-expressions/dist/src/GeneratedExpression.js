"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const util_1 = __importDefault(require("util"));
class GeneratedExpression {
    constructor(expressionTemplate, parameterTypes) {
        this.expressionTemplate = expressionTemplate;
        this.parameterTypes = parameterTypes;
    }
    get source() {
        return util_1.default.format(this.expressionTemplate, ...this.parameterTypes.map((t) => t.name));
    }
    /**
     * Returns an array of parameter names to use in generated function/method signatures
     *
     * @returns {ReadonlyArray.<String>}
     */
    get parameterNames() {
        const usageByTypeName = {};
        return this.parameterTypes.map((t) => getParameterName(t.name, usageByTypeName));
    }
}
exports.default = GeneratedExpression;
function getParameterName(typeName, usageByTypeName) {
    let count = usageByTypeName[typeName];
    count = count ? count + 1 : 1;
    usageByTypeName[typeName] = count;
    return count === 1 ? typeName : `${typeName}${count}`;
}
//# sourceMappingURL=GeneratedExpression.js.map