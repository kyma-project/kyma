import ParameterType from './ParameterType';
import GeneratedExpression from './GeneratedExpression';
export default class CucumberExpressionGenerator {
    private readonly parameterTypes;
    constructor(parameterTypes: () => Iterable<ParameterType<any>>);
    generateExpressions(text: string): readonly GeneratedExpression[];
    /**
     * @deprecated
     */
    generateExpression(text: string): GeneratedExpression;
    private createParameterTypeMatchers;
    private static createParameterTypeMatchers2;
}
//# sourceMappingURL=CucumberExpressionGenerator.d.ts.map