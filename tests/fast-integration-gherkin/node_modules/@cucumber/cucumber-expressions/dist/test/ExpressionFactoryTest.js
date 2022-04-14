"use strict";
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    Object.defineProperty(o, k2, { enumerable: true, get: function() { return m[k]; } });
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || function (mod) {
    if (mod && mod.__esModule) return mod;
    var result = {};
    if (mod != null) for (var k in mod) if (k !== "default" && Object.prototype.hasOwnProperty.call(mod, k)) __createBinding(result, mod, k);
    __setModuleDefault(result, mod);
    return result;
};
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const assert = __importStar(require("assert"));
const ExpressionFactory_1 = __importDefault(require("../src/ExpressionFactory"));
const ParameterTypeRegistry_1 = __importDefault(require("../src/ParameterTypeRegistry"));
const RegularExpression_1 = __importDefault(require("../src/RegularExpression"));
const CucumberExpression_1 = __importDefault(require("../src/CucumberExpression"));
describe('ExpressionFactory', () => {
    let expressionFactory;
    beforeEach(() => {
        expressionFactory = new ExpressionFactory_1.default(new ParameterTypeRegistry_1.default());
    });
    it('creates a RegularExpression', () => {
        assert.strictEqual(expressionFactory.createExpression(/x/).constructor, RegularExpression_1.default);
    });
    it('creates a CucumberExpression', () => {
        assert.strictEqual(expressionFactory.createExpression('x').constructor, CucumberExpression_1.default);
    });
});
//# sourceMappingURL=ExpressionFactoryTest.js.map