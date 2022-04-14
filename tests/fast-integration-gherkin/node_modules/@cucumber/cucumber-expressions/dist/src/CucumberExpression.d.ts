import ParameterTypeRegistry from './ParameterTypeRegistry';
import Argument from './Argument';
import Expression from './Expression';
export default class CucumberExpression implements Expression {
    private readonly expression;
    private readonly parameterTypeRegistry;
    private readonly parameterTypes;
    private readonly treeRegexp;
    /**
     * @param expression
     * @param parameterTypeRegistry
     */
    constructor(expression: string, parameterTypeRegistry: ParameterTypeRegistry);
    private rewriteToRegex;
    private static escapeRegex;
    private rewriteOptional;
    private rewriteAlternation;
    private rewriteAlternative;
    private rewriteParameter;
    private rewriteExpression;
    private assertNotEmpty;
    private assertNoParameters;
    private assertNoOptionals;
    match(text: string): readonly Argument<any>[];
    get regexp(): RegExp;
    get source(): string;
}
//# sourceMappingURL=CucumberExpression.d.ts.map