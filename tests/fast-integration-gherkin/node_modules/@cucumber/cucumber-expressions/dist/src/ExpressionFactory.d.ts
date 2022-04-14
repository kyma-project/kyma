import Expression from './Expression';
import ParameterTypeRegistry from './ParameterTypeRegistry';
export default class ExpressionFactory {
    private readonly parameterTypeRegistry;
    constructor(parameterTypeRegistry: ParameterTypeRegistry);
    createExpression(expression: string | RegExp): Expression;
}
//# sourceMappingURL=ExpressionFactory.d.ts.map