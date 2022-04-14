"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const CucumberExpression_1 = __importDefault(require("./CucumberExpression"));
const RegularExpression_1 = __importDefault(require("./RegularExpression"));
class ExpressionFactory {
    constructor(parameterTypeRegistry) {
        this.parameterTypeRegistry = parameterTypeRegistry;
    }
    createExpression(expression) {
        return typeof expression === 'string'
            ? new CucumberExpression_1.default(expression, this.parameterTypeRegistry)
            : new RegularExpression_1.default(expression, this.parameterTypeRegistry);
    }
}
exports.default = ExpressionFactory;
//# sourceMappingURL=ExpressionFactory.js.map