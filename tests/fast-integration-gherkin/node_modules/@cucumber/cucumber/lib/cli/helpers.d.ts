/// <reference types="node" />
import { EventEmitter } from 'events';
import PickleFilter from '../pickle_filter';
import { EventDataCollector } from '../formatter/helpers';
import { Readable } from 'stream';
import { IdGenerator } from '@cucumber/messages';
import { ISupportCodeLibrary } from '../support_code_library_builder/types';
export interface IGetExpandedArgvRequest {
    argv: string[];
    cwd: string;
}
export declare function getExpandedArgv({ argv, cwd, }: IGetExpandedArgvRequest): Promise<string[]>;
interface IParseGherkinMessageStreamRequest {
    cwd: string;
    eventBroadcaster: EventEmitter;
    eventDataCollector: EventDataCollector;
    gherkinMessageStream: Readable;
    order: string;
    pickleFilter: PickleFilter;
}
export declare function parseGherkinMessageStream({ cwd, eventBroadcaster, eventDataCollector, gherkinMessageStream, order, pickleFilter, }: IParseGherkinMessageStreamRequest): Promise<string[]>;
export declare function orderPickleIds(pickleIds: string[], order: string): void;
export declare function emitMetaMessage(eventBroadcaster: EventEmitter): Promise<void>;
export declare function emitSupportCodeMessages({ eventBroadcaster, supportCodeLibrary, newId, }: {
    eventBroadcaster: EventEmitter;
    supportCodeLibrary: ISupportCodeLibrary;
    newId: IdGenerator.NewId;
}): void;
export {};
