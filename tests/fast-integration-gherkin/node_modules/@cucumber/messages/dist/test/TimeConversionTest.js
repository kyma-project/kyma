"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const assert_1 = __importDefault(require("assert"));
const src_1 = require("../src");
const TimeConversion_1 = require("../src/TimeConversion");
const { durationToMilliseconds, millisecondsSinceEpochToTimestamp, millisecondsToDuration, timestampToMillisecondsSinceEpoch, } = src_1.TimeConversion;
describe('TimeConversion', () => {
    it('converts legacy string seconds', () => {
        const duration = {
            // @ts-ignore
            seconds: '3',
            nanos: 40000,
        };
        const millis = durationToMilliseconds(duration);
        assert_1.default.strictEqual(millis, 3000.04);
    });
    it('converts to and from milliseconds since epoch', () => {
        const millisecondsSinceEpoch = Date.now();
        const timestamp = millisecondsSinceEpochToTimestamp(millisecondsSinceEpoch);
        const jsEpochMillisAgain = timestampToMillisecondsSinceEpoch(timestamp);
        assert_1.default.strictEqual(jsEpochMillisAgain, millisecondsSinceEpoch);
    });
    it('converts to and from milliseconds duration', () => {
        const durationInMilliseconds = 1234;
        const duration = millisecondsToDuration(durationInMilliseconds);
        const durationMillisAgain = durationToMilliseconds(duration);
        assert_1.default.strictEqual(durationMillisAgain, durationInMilliseconds);
    });
    it('converts to and from milliseconds duration (with decimal places)', () => {
        const durationInMilliseconds = 3.000161;
        const duration = millisecondsToDuration(durationInMilliseconds);
        const durationMillisAgain = durationToMilliseconds(duration);
        assert_1.default.strictEqual(durationMillisAgain, durationInMilliseconds);
    });
    it('adds durations (nanos only)', () => {
        const durationA = millisecondsToDuration(100);
        const durationB = millisecondsToDuration(200);
        const sumDuration = TimeConversion_1.addDurations(durationA, durationB);
        assert_1.default.deepStrictEqual(sumDuration, { seconds: 0, nanos: 3e8 });
    });
    it('adds durations (seconds only)', () => {
        const durationA = millisecondsToDuration(1000);
        const durationB = millisecondsToDuration(2000);
        const sumDuration = TimeConversion_1.addDurations(durationA, durationB);
        assert_1.default.deepStrictEqual(sumDuration, { seconds: 3, nanos: 0 });
    });
    it('adds durations (seconds and nanos)', () => {
        const durationA = millisecondsToDuration(1500);
        const durationB = millisecondsToDuration(1600);
        const sumDuration = TimeConversion_1.addDurations(durationA, durationB);
        assert_1.default.deepStrictEqual(sumDuration, { seconds: 3, nanos: 1e8 });
    });
    it('adds durations (seconds and nanos) with legacy string seconds', () => {
        const durationA = millisecondsToDuration(1500);
        // @ts-ignore
        durationA.seconds = String(durationA.seconds);
        const durationB = millisecondsToDuration(1600);
        // @ts-ignore
        durationB.seconds = String(durationB.seconds);
        const sumDuration = TimeConversion_1.addDurations(durationA, durationB);
        assert_1.default.deepStrictEqual(sumDuration, { seconds: 3, nanos: 1e8 });
    });
});
//# sourceMappingURL=TimeConversionTest.js.map