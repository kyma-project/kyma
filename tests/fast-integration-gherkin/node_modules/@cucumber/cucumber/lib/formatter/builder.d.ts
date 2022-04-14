/// <reference types="node" />
import StepDefinitionSnippetBuilder from './step_definition_snippet_builder';
import { ISupportCodeLibrary } from '../support_code_library_builder/types';
import Formatter, { IFormatterCleanupFn, IFormatterLogFn } from '.';
import { EventEmitter } from 'events';
import EventDataCollector from './helpers/event_data_collector';
import { Writable as WritableStream } from 'stream';
import { IParsedArgvFormatOptions } from '../cli/argv_parser';
import { SnippetInterface } from './step_definition_snippet_builder/snippet_syntax';
interface IGetStepDefinitionSnippetBuilderOptions {
    cwd: string;
    snippetInterface?: SnippetInterface;
    snippetSyntax?: string;
    supportCodeLibrary: ISupportCodeLibrary;
}
export interface IBuildOptions {
    cwd: string;
    eventBroadcaster: EventEmitter;
    eventDataCollector: EventDataCollector;
    log: IFormatterLogFn;
    parsedArgvOptions: IParsedArgvFormatOptions;
    stream: WritableStream;
    cleanup: IFormatterCleanupFn;
    supportCodeLibrary: ISupportCodeLibrary;
}
declare const FormatterBuilder: {
    build(type: string, options: IBuildOptions): Formatter;
    getConstructorByType(type: string, cwd: string): typeof Formatter;
    getStepDefinitionSnippetBuilder({ cwd, snippetInterface, snippetSyntax, supportCodeLibrary, }: IGetStepDefinitionSnippetBuilderOptions): StepDefinitionSnippetBuilder;
    loadCustomFormatter(customFormatterPath: string, cwd: string): any;
};
export default FormatterBuilder;
