import * as messages from '@cucumber/messages';
import { IColorFns } from '../get_color_fns';
import StepDefinitionSnippetBuilder from '../step_definition_snippet_builder';
import { ISupportCodeLibrary } from '../../support_code_library_builder/types';
import { ITestCaseAttempt } from './event_data_collector';
export declare function isFailure(result: messages.TestStepResult): boolean;
export declare function isWarning(result: messages.TestStepResult): boolean;
export declare function isIssue(result: messages.TestStepResult): boolean;
export interface IFormatIssueRequest {
    colorFns: IColorFns;
    cwd: string;
    number: number;
    snippetBuilder: StepDefinitionSnippetBuilder;
    testCaseAttempt: ITestCaseAttempt;
    supportCodeLibrary: ISupportCodeLibrary;
}
export declare function formatIssue({ colorFns, cwd, number, snippetBuilder, testCaseAttempt, supportCodeLibrary, }: IFormatIssueRequest): string;
export declare function formatUndefinedParameterTypes(undefinedParameterTypes: messages.UndefinedParameterType[]): string;
export declare function formatUndefinedParameterType(parameterType: messages.UndefinedParameterType): string;
