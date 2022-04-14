"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const GeneratedExpression_1 = __importDefault(require("./GeneratedExpression"));
// 256 generated expressions ought to be enough for anybody
const MAX_EXPRESSIONS = 256;
class CombinatorialGeneratedExpressionFactory {
    constructor(expressionTemplate, parameterTypeCombinations) {
        this.expressionTemplate = expressionTemplate;
        this.parameterTypeCombinations = parameterTypeCombinations;
        this.expressionTemplate = expressionTemplate;
    }
    generateExpressions() {
        const generatedExpressions = [];
        this.generatePermutations(generatedExpressions, 0, []);
        return generatedExpressions;
    }
    generatePermutations(generatedExpressions, depth, currentParameterTypes) {
        if (generatedExpressions.length >= MAX_EXPRESSIONS) {
            return;
        }
        if (depth === this.parameterTypeCombinations.length) {
            generatedExpressions.push(new GeneratedExpression_1.default(this.expressionTemplate, currentParameterTypes));
            return;
        }
        // tslint:disable-next-line:prefer-for-of
        for (let i = 0; i < this.parameterTypeCombinations[depth].length; ++i) {
            // Avoid recursion if no elements can be added.
            if (generatedExpressions.length >= MAX_EXPRESSIONS) {
                return;
            }
            const newCurrentParameterTypes = currentParameterTypes.slice(); // clone
            newCurrentParameterTypes.push(this.parameterTypeCombinations[depth][i]);
            this.generatePermutations(generatedExpressions, depth + 1, newCurrentParameterTypes);
        }
    }
}
exports.default = CombinatorialGeneratedExpressionFactory;
//# sourceMappingURL=CombinatorialGeneratedExpressionFactory.js.map