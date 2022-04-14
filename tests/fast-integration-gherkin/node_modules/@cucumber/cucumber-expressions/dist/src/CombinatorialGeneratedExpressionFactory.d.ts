import GeneratedExpression from './GeneratedExpression';
import ParameterType from './ParameterType';
export default class CombinatorialGeneratedExpressionFactory {
    private readonly expressionTemplate;
    private readonly parameterTypeCombinations;
    constructor(expressionTemplate: string, parameterTypeCombinations: Array<Array<ParameterType<any>>>);
    generateExpressions(): readonly GeneratedExpression[];
    private generatePermutations;
}
//# sourceMappingURL=CombinatorialGeneratedExpressionFactory.d.ts.map