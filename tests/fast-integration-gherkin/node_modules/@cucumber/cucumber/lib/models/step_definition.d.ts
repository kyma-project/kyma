import Definition, { IDefinition, IGetInvocationDataRequest, IGetInvocationDataResponse, IStepDefinitionParameters } from './definition';
import { Expression } from '@cucumber/cucumber-expressions';
export default class StepDefinition extends Definition implements IDefinition {
    readonly pattern: string | RegExp;
    readonly expression: Expression;
    constructor(data: IStepDefinitionParameters);
    getInvocationParameters({ step, world, }: IGetInvocationDataRequest): Promise<IGetInvocationDataResponse>;
    matchesStepName(stepName: string): boolean;
}
