"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.addDurations = exports.durationToMilliseconds = exports.timestampToMillisecondsSinceEpoch = exports.millisecondsToDuration = exports.millisecondsSinceEpochToTimestamp = void 0;
const MILLISECONDS_PER_SECOND = 1e3;
const NANOSECONDS_PER_MILLISECOND = 1e6;
const NANOSECONDS_PER_SECOND = 1e9;
function millisecondsSinceEpochToTimestamp(millisecondsSinceEpoch) {
    return toSecondsAndNanos(millisecondsSinceEpoch);
}
exports.millisecondsSinceEpochToTimestamp = millisecondsSinceEpochToTimestamp;
function millisecondsToDuration(durationInMilliseconds) {
    return toSecondsAndNanos(durationInMilliseconds);
}
exports.millisecondsToDuration = millisecondsToDuration;
function timestampToMillisecondsSinceEpoch(timestamp) {
    const { seconds, nanos } = timestamp;
    return toMillis(seconds, nanos);
}
exports.timestampToMillisecondsSinceEpoch = timestampToMillisecondsSinceEpoch;
function durationToMilliseconds(duration) {
    const { seconds, nanos } = duration;
    return toMillis(seconds, nanos);
}
exports.durationToMilliseconds = durationToMilliseconds;
function addDurations(durationA, durationB) {
    let seconds = +durationA.seconds + +durationB.seconds;
    let nanos = durationA.nanos + durationB.nanos;
    if (nanos >= NANOSECONDS_PER_SECOND) {
        seconds += 1;
        nanos -= NANOSECONDS_PER_SECOND;
    }
    return { seconds, nanos };
}
exports.addDurations = addDurations;
function toSecondsAndNanos(milliseconds) {
    const seconds = Math.floor(milliseconds / MILLISECONDS_PER_SECOND);
    const nanos = Math.floor((milliseconds % MILLISECONDS_PER_SECOND) * NANOSECONDS_PER_MILLISECOND);
    return { seconds, nanos };
}
function toMillis(seconds, nanos) {
    const secondMillis = +seconds * MILLISECONDS_PER_SECOND;
    const nanoMillis = nanos / NANOSECONDS_PER_MILLISECOND;
    return secondMillis + nanoMillis;
}
//# sourceMappingURL=TimeConversion.js.map