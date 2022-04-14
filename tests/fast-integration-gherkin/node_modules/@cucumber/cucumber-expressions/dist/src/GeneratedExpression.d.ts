import ParameterType from './ParameterType';
export default class GeneratedExpression {
    private readonly expressionTemplate;
    readonly parameterTypes: readonly ParameterType<any>[];
    constructor(expressionTemplate: string, parameterTypes: readonly ParameterType<any>[]);
    get source(): string;
    /**
     * Returns an array of parameter names to use in generated function/method signatures
     *
     * @returns {ReadonlyArray.<String>}
     */
    get parameterNames(): readonly string[];
}
//# sourceMappingURL=GeneratedExpression.d.ts.map