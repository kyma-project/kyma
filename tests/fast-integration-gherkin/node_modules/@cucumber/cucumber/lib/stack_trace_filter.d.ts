/// <reference types="node" />
import CallSite = NodeJS.CallSite;
export declare function isFileNameInCucumber(fileName: string): boolean;
export default class StackTraceFilter {
    private currentFilter;
    filter(): void;
    isErrorInCucumber(frames: CallSite[]): boolean;
    isFrameInCucumber(frame: CallSite): boolean;
    isFrameInNode(frame: CallSite): boolean;
    unfilter(): void;
}
