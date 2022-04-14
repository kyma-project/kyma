export interface IDefinitionConfig {
    code: any;
    line: number;
    uri: string;
}
export interface IValidateNoGeneratorFunctionsRequest {
    cwd: string;
    definitionConfigs: IDefinitionConfig[];
}
export declare function validateNoGeneratorFunctions({ cwd, definitionConfigs, }: IValidateNoGeneratorFunctionsRequest): void;
