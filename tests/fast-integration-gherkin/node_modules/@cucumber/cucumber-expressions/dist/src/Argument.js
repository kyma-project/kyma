"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const CucumberExpressionError_1 = __importDefault(require("./CucumberExpressionError"));
class Argument {
    constructor(group, parameterType) {
        this.group = group;
        this.parameterType = parameterType;
        this.group = group;
        this.parameterType = parameterType;
    }
    static build(treeRegexp, text, parameterTypes) {
        const group = treeRegexp.match(text);
        if (!group) {
            return null;
        }
        const argGroups = group.children;
        if (argGroups.length !== parameterTypes.length) {
            throw new CucumberExpressionError_1.default(`Expression ${treeRegexp.regexp} has ${argGroups.length} capture groups (${argGroups.map((g) => g.value)}), but there were ${parameterTypes.length} parameter types (${parameterTypes.map((p) => p.name)})`);
        }
        return parameterTypes.map((parameterType, i) => new Argument(argGroups[i], parameterType));
    }
    /**
     * Get the value returned by the parameter type's transformer function.
     *
     * @param thisObj the object in which the transformer function is applied.
     */
    getValue(thisObj) {
        const groupValues = this.group ? this.group.values : null;
        return this.parameterType.transform(thisObj, groupValues);
    }
    getParameterType() {
        return this.parameterType;
    }
}
exports.default = Argument;
module.exports = Argument;
//# sourceMappingURL=Argument.js.map