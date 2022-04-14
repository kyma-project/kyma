/// <reference types="node" />
import { Readable } from 'stream';
import * as messages from '@cucumber/messages';
export interface IAttachmentMedia {
    encoding: messages.AttachmentContentEncoding;
    contentType: string;
}
export interface IAttachment {
    data: string;
    media: IAttachmentMedia;
}
export declare type IAttachFunction = (attachment: IAttachment) => void;
export declare type ICreateAttachment = (data: Buffer | Readable | string, mediaType?: string, callback?: () => void) => void | Promise<void>;
export declare type ICreateLog = (text: string) => void | Promise<void>;
export default class AttachmentManager {
    private readonly onAttachment;
    constructor(onAttachment: IAttachFunction);
    log(text: string): void | Promise<void>;
    create(data: Buffer | Readable | string, mediaType?: string, callback?: () => void): void | Promise<void>;
    createBufferAttachment(data: Buffer, mediaType: string): void;
    createStreamAttachment(data: Readable, mediaType: string, callback: () => void): void | Promise<void>;
    createStringAttachment(data: string, media: IAttachmentMedia): void;
}
