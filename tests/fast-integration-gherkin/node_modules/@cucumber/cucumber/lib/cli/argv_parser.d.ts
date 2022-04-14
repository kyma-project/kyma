import { SnippetInterface } from '../formatter/step_definition_snippet_builder/snippet_syntax';
export interface IParsedArgvFormatRerunOptions {
    separator?: string;
}
export interface IParsedArgvFormatOptions {
    colorsEnabled?: boolean;
    rerun?: IParsedArgvFormatRerunOptions;
    snippetInterface?: SnippetInterface;
    snippetSyntax?: string;
    [customKey: string]: any;
}
export interface IParsedArgvOptions {
    backtrace: boolean;
    dryRun: boolean;
    exit: boolean;
    failFast: boolean;
    format: string[];
    formatOptions: IParsedArgvFormatOptions;
    i18nKeywords: string;
    i18nLanguages: boolean;
    language: string;
    name: string[];
    order: string;
    parallel: number;
    predictableIds: boolean;
    profile: string[];
    publish: boolean;
    publishQuiet: boolean;
    require: string[];
    requireModule: string[];
    retry: number;
    retryTagFilter: string;
    strict: boolean;
    tags: string;
    worldParameters: object;
}
export interface IParsedArgv {
    args: string[];
    options: IParsedArgvOptions;
}
declare const ArgvParser: {
    collect<T>(val: T, memo: T[]): T[];
    mergeJson(option: string): (str: string, memo: object) => object;
    mergeTags(value: string, memo: string): string;
    validateCountOption(value: string, optionName: string): number;
    validateLanguage(value: string): string;
    validateRetryOptions(options: IParsedArgvOptions): void;
    parse(argv: string[]): IParsedArgv;
    lint(fullArgv: string[]): void;
};
export default ArgvParser;
