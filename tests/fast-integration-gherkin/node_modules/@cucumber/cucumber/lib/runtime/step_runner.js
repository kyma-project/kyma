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
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.run = void 0;
const lodash_1 = __importDefault(require("lodash"));
const time_1 = __importDefault(require("../time"));
const user_code_runner_1 = __importDefault(require("../user_code_runner"));
const messages = __importStar(require("@cucumber/messages"));
const assertion_error_formatter_1 = require("assertion-error-formatter");
const value_checker_1 = require("../value_checker");
const { beginTiming, endTiming } = time_1.default;
async function run({ defaultTimeout, hookParameter, step, stepDefinition, world, }) {
    beginTiming();
    let error, result, invocationData;
    try {
        invocationData = await stepDefinition.getInvocationParameters({
            hookParameter,
            step,
            world,
        });
    }
    catch (err) {
        error = err;
    }
    if (value_checker_1.doesNotHaveValue(error)) {
        const timeoutInMilliseconds = value_checker_1.valueOrDefault(stepDefinition.options.timeout, defaultTimeout);
        if (lodash_1.default.includes(invocationData.validCodeLengths, stepDefinition.code.length)) {
            const data = await user_code_runner_1.default.run({
                argsArray: invocationData.parameters,
                fn: stepDefinition.code,
                thisArg: world,
                timeoutInMilliseconds,
            });
            error = data.error;
            result = data.result;
        }
        else {
            error = invocationData.getInvalidCodeLengthMessage();
        }
    }
    const duration = messages.TimeConversion.millisecondsToDuration(endTiming());
    let status;
    let message;
    if (result === 'skipped') {
        status = messages.TestStepResultStatus.SKIPPED;
    }
    else if (result === 'pending') {
        status = messages.TestStepResultStatus.PENDING;
    }
    else if (value_checker_1.doesHaveValue(error)) {
        message = assertion_error_formatter_1.format(error);
        status = messages.TestStepResultStatus.FAILED;
    }
    else {
        status = messages.TestStepResultStatus.PASSED;
    }
    return {
        duration,
        status,
        message,
        willBeRetried: false,
    };
}
exports.run = run;
exports.default = { run };
//# sourceMappingURL=step_runner.js.map