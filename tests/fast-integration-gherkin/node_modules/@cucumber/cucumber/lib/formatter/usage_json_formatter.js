"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const helpers_1 = require("./helpers");
const _1 = __importDefault(require("./"));
const value_checker_1 = require("../value_checker");
class UsageJsonFormatter extends _1.default {
    constructor(options) {
        super(options);
        options.eventBroadcaster.on('envelope', (envelope) => {
            if (value_checker_1.doesHaveValue(envelope.testRunFinished)) {
                this.logUsage();
            }
        });
    }
    logUsage() {
        const usage = helpers_1.getUsage({
            cwd: this.cwd,
            stepDefinitions: this.supportCodeLibrary.stepDefinitions,
            eventDataCollector: this.eventDataCollector,
        });
        this.log(JSON.stringify(usage, this.replacer, 2));
    }
    replacer(key, value) {
        if (key === 'seconds') {
            return parseInt(value);
        }
        return value;
    }
}
exports.default = UsageJsonFormatter;
//# sourceMappingURL=usage_json_formatter.js.map