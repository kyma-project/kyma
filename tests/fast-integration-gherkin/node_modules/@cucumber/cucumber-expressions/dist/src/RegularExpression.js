"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const Argument_1 = __importDefault(require("./Argument"));
const TreeRegexp_1 = __importDefault(require("./TreeRegexp"));
const ParameterType_1 = __importDefault(require("./ParameterType"));
class RegularExpression {
    constructor(regexp, parameterTypeRegistry) {
        this.regexp = regexp;
        this.parameterTypeRegistry = parameterTypeRegistry;
        this.treeRegexp = new TreeRegexp_1.default(regexp);
    }
    match(text) {
        const parameterTypes = this.treeRegexp.groupBuilder.children.map((groupBuilder) => {
            const parameterTypeRegexp = groupBuilder.source;
            return (this.parameterTypeRegistry.lookupByRegexp(parameterTypeRegexp, this.regexp, text) ||
                new ParameterType_1.default(null, parameterTypeRegexp, String, (s) => (s === undefined ? null : s), false, false));
        });
        return Argument_1.default.build(this.treeRegexp, text, parameterTypes);
    }
    get source() {
        return this.regexp.source;
    }
}
exports.default = RegularExpression;
//# sourceMappingURL=RegularExpression.js.map