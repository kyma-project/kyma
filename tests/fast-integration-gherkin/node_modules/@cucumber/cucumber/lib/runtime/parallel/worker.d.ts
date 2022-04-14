import { ICoordinatorReport, IWorkerCommand, IWorkerCommandInitialize, IWorkerCommandRun } from './command_types';
import TestRunHookDefinition from '../../models/test_run_hook_definition';
declare type IExitFunction = (exitCode: number, error?: Error, message?: string) => void;
declare type IMessageSender = (command: ICoordinatorReport) => void;
export default class Worker {
    private readonly cwd;
    private readonly exit;
    private readonly id;
    private readonly eventBroadcaster;
    private filterStacktraces;
    private readonly newId;
    private readonly sendMessage;
    private readonly stackTraceFilter;
    private supportCodeLibrary;
    private worldParameters;
    private options;
    constructor({ cwd, exit, id, sendMessage, }: {
        cwd: string;
        exit: IExitFunction;
        id: string;
        sendMessage: IMessageSender;
    });
    initialize({ filterStacktraces, supportCodeRequiredModules, supportCodePaths, supportCodeIds, options, }: IWorkerCommandInitialize): Promise<void>;
    finalize(): Promise<void>;
    receiveMessage(message: IWorkerCommand): Promise<void>;
    runTestCase({ gherkinDocument, pickle, testCase, elapsed, retries, skip, }: IWorkerCommandRun): Promise<void>;
    runTestRunHooks(testRunHookDefinitions: TestRunHookDefinition[], name: string): Promise<void>;
}
export {};
