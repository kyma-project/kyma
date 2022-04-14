import * as messages from '@cucumber/messages';
import { Duration } from 'durations';
export interface ITestRunStopwatch {
    from: (duration: Duration) => ITestRunStopwatch;
    start: () => ITestRunStopwatch;
    stop: () => ITestRunStopwatch;
    duration: () => Duration;
    timestamp: () => messages.Timestamp;
}
export declare class RealTestRunStopwatch implements ITestRunStopwatch {
    private readonly stopwatch;
    private base;
    from(duration: Duration): ITestRunStopwatch;
    start(): ITestRunStopwatch;
    stop(): ITestRunStopwatch;
    duration(): Duration;
    timestamp(): messages.Timestamp;
}
export declare class PredictableTestRunStopwatch implements ITestRunStopwatch {
    private count;
    private base;
    from(duration: Duration): ITestRunStopwatch;
    start(): ITestRunStopwatch;
    stop(): ITestRunStopwatch;
    duration(): Duration;
    timestamp(): messages.Timestamp;
    private convertToTimestamp;
}
