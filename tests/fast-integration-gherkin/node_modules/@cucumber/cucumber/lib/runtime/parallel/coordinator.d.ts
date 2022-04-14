/// <reference types="node" />
import { ChildProcess } from 'child_process';
import * as messages from '@cucumber/messages';
import { EventEmitter } from 'events';
import { EventDataCollector } from '../../formatter/helpers';
import { IRuntimeOptions } from '..';
import { ISupportCodeLibrary } from '../../support_code_library_builder/types';
import { ICoordinatorReport } from './command_types';
import { IdGenerator } from '@cucumber/messages';
export interface INewCoordinatorOptions {
    cwd: string;
    eventBroadcaster: EventEmitter;
    eventDataCollector: EventDataCollector;
    options: IRuntimeOptions;
    newId: IdGenerator.NewId;
    pickleIds: string[];
    supportCodeLibrary: ISupportCodeLibrary;
    supportCodePaths: string[];
    supportCodeRequiredModules: string[];
}
interface IWorker {
    closed: boolean;
    process: ChildProcess;
}
export default class Coordinator {
    private readonly cwd;
    private readonly eventBroadcaster;
    private readonly eventDataCollector;
    private readonly stopwatch;
    private onFinish;
    private nextPickleIdIndex;
    private readonly options;
    private readonly newId;
    private readonly pickleIds;
    private assembledTestCases;
    private workers;
    private readonly supportCodeLibrary;
    private readonly supportCodePaths;
    private readonly supportCodeRequiredModules;
    private success;
    constructor({ cwd, eventBroadcaster, eventDataCollector, pickleIds, options, newId, supportCodeLibrary, supportCodePaths, supportCodeRequiredModules, }: INewCoordinatorOptions);
    parseWorkerMessage(worker: IWorker, message: ICoordinatorReport): void;
    startWorker(id: string, total: number): void;
    onWorkerProcessClose(exitCode: number): void;
    parseTestCaseResult(testCaseFinished: messages.TestCaseFinished): void;
    run(numberOfWorkers: number): Promise<boolean>;
    giveWork(worker: IWorker): void;
    shouldCauseFailure(status: messages.TestStepResultStatus): boolean;
}
export {};
