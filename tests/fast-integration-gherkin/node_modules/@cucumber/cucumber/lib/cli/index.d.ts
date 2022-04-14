/// <reference types="node" />
import { EventDataCollector } from '../formatter/helpers';
import { IConfiguration, IConfigurationFormat } from './configuration_builder';
import { EventEmitter } from 'events';
import { IdGenerator } from '@cucumber/messages';
import { IFormatterStream } from '../formatter';
import { ISupportCodeLibrary } from '../support_code_library_builder/types';
import { IParsedArgvFormatOptions } from './argv_parser';
export interface ICliRunResult {
    shouldExitImmediately: boolean;
    success: boolean;
}
interface IInitializeFormattersRequest {
    eventBroadcaster: EventEmitter;
    eventDataCollector: EventDataCollector;
    formatOptions: IParsedArgvFormatOptions;
    formats: IConfigurationFormat[];
    supportCodeLibrary: ISupportCodeLibrary;
}
interface IGetSupportCodeLibraryRequest {
    newId: IdGenerator.NewId;
    supportCodeRequiredModules: string[];
    supportCodePaths: string[];
}
export default class Cli {
    private readonly argv;
    private readonly cwd;
    private readonly stdout;
    constructor({ argv, cwd, stdout, }: {
        argv: string[];
        cwd: string;
        stdout: IFormatterStream;
    });
    getConfiguration(): Promise<IConfiguration>;
    initializeFormatters({ eventBroadcaster, eventDataCollector, formatOptions, formats, supportCodeLibrary, }: IInitializeFormattersRequest): Promise<() => Promise<void>>;
    getSupportCodeLibrary({ newId, supportCodeRequiredModules, supportCodePaths, }: IGetSupportCodeLibraryRequest): ISupportCodeLibrary;
    run(): Promise<ICliRunResult>;
}
export {};
