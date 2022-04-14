"use strict";
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    Object.defineProperty(o, k2, { enumerable: true, get: function() { return m[k]; } });
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || function (mod) {
    if (mod && mod.__esModule) return mod;
    var result = {};
    if (mod != null) for (var k in mod) if (k !== "default" && Object.prototype.hasOwnProperty.call(mod, k)) __createBinding(result, mod, k);
    __setModuleDefault(result, mod);
    return result;
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.PredictableTestRunStopwatch = exports.RealTestRunStopwatch = void 0;
const messages = __importStar(require("@cucumber/messages"));
const durations_1 = require("durations");
class RealTestRunStopwatch {
    constructor() {
        this.stopwatch = durations_1.stopwatch();
        this.base = null;
    }
    from(duration) {
        this.base = duration;
        return this;
    }
    start() {
        this.stopwatch.start();
        return this;
    }
    stop() {
        this.stopwatch.stop();
        return this;
    }
    duration() {
        const current = this.stopwatch.duration();
        if (this.base !== null) {
            return durations_1.duration(this.base.nanos() + current.nanos());
        }
        return current;
    }
    timestamp() {
        return messages.TimeConversion.millisecondsSinceEpochToTimestamp(Date.now());
    }
}
exports.RealTestRunStopwatch = RealTestRunStopwatch;
class PredictableTestRunStopwatch {
    constructor() {
        this.count = 0;
        this.base = null;
    }
    from(duration) {
        this.base = duration;
        return this;
    }
    start() {
        return this;
    }
    stop() {
        return this;
    }
    duration() {
        const current = durations_1.duration(this.count * 1000000);
        if (this.base !== null) {
            return durations_1.duration(this.base.nanos() + current.nanos());
        }
        return current;
    }
    timestamp() {
        const fakeTimestamp = this.convertToTimestamp(this.duration());
        this.count++;
        return fakeTimestamp;
    }
    // TODO: Remove. It's impossible to convert timestamps to durations and vice-versa
    convertToTimestamp(duration) {
        const seconds = Math.floor(duration.seconds());
        const nanos = Math.floor((duration.seconds() - seconds) * 1000000000);
        return {
            seconds,
            nanos,
        };
    }
}
exports.PredictableTestRunStopwatch = PredictableTestRunStopwatch;
//# sourceMappingURL=stopwatch.js.map