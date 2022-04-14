/// <reference types="node" />
import { EventDataCollector } from '../formatter/helpers';
import { IdGenerator } from '@cucumber/messages';
import * as messages from '@cucumber/messages';
import { EventEmitter } from 'events';
import { ISupportCodeLibrary } from '../support_code_library_builder/types';
import TestRunHookDefinition from '../models/test_run_hook_definition';
export interface INewRuntimeOptions {
    eventBroadcaster: EventEmitter;
    eventDataCollector: EventDataCollector;
    newId: IdGenerator.NewId;
    options: IRuntimeOptions;
    pickleIds: string[];
    supportCodeLibrary: ISupportCodeLibrary;
}
export interface IRuntimeOptions {
    dryRun: boolean;
    predictableIds: boolean;
    failFast: boolean;
    filterStacktraces: boolean;
    retry: number;
    retryTagFilter: string;
    strict: boolean;
    worldParameters: any;
}
export default class Runtime {
    private readonly eventBroadcaster;
    private readonly eventDataCollector;
    private readonly stopwatch;
    private readonly newId;
    private readonly options;
    private readonly pickleIds;
    private readonly stackTraceFilter;
    private readonly supportCodeLibrary;
    private success;
    constructor({ eventBroadcaster, eventDataCollector, newId, options, pickleIds, supportCodeLibrary, }: INewRuntimeOptions);
    runTestRunHooks(definitions: TestRunHookDefinition[], name: string): Promise<void>;
    runTestCase(pickleId: string, testCase: messages.TestCase): Promise<void>;
    start(): Promise<boolean>;
    shouldCauseFailure(status: messages.TestStepResultStatus): boolean;
}
