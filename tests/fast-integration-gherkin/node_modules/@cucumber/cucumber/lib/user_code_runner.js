"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const bluebird_1 = __importDefault(require("bluebird"));
const time_1 = __importDefault(require("./time"));
const uncaught_exception_manager_1 = __importDefault(require("./uncaught_exception_manager"));
const util_1 = __importDefault(require("util"));
const value_checker_1 = require("./value_checker");
const UserCodeRunner = {
    async run({ argsArray, thisArg, fn, timeoutInMilliseconds, }) {
        const callbackPromise = new Promise((resolve, reject) => {
            argsArray.push((error, result) => {
                if (value_checker_1.doesHaveValue(error)) {
                    reject(error);
                }
                else {
                    resolve(result);
                }
            });
        });
        let fnReturn;
        try {
            fnReturn = fn.apply(thisArg, argsArray);
        }
        catch (e) {
            const error = e instanceof Error ? e : util_1.default.format(e);
            return { error };
        }
        const racingPromises = [];
        const callbackInterface = fn.length === argsArray.length;
        const promiseInterface = value_checker_1.doesHaveValue(fnReturn) && typeof fnReturn.then === 'function';
        if (callbackInterface && promiseInterface) {
            return {
                error: new Error('function uses multiple asynchronous interfaces: callback and promise\n' +
                    'to use the callback interface: do not return a promise\n' +
                    'to use the promise interface: remove the last argument to the function'),
            };
        }
        else if (callbackInterface) {
            racingPromises.push(callbackPromise);
        }
        else if (promiseInterface) {
            racingPromises.push(fnReturn);
        }
        else {
            return { result: fnReturn };
        }
        let exceptionHandler;
        const uncaughtExceptionPromise = new Promise((resolve, reject) => {
            exceptionHandler = reject;
            uncaught_exception_manager_1.default.registerHandler(exceptionHandler);
        });
        racingPromises.push(uncaughtExceptionPromise);
        let timeoutId;
        if (timeoutInMilliseconds >= 0) {
            const timeoutPromise = new Promise((resolve, reject) => {
                timeoutId = time_1.default.setTimeout(() => {
                    const timeoutMessage = 'function timed out, ensure the ' +
                        (callbackInterface ? 'callback is executed' : 'promise resolves') +
                        ` within ${timeoutInMilliseconds.toString()} milliseconds`;
                    reject(new Error(timeoutMessage));
                }, timeoutInMilliseconds);
            });
            racingPromises.push(timeoutPromise);
        }
        let error, result;
        try {
            result = await bluebird_1.default.race(racingPromises);
        }
        catch (e) {
            if (e instanceof Error) {
                error = e;
            }
            else if (value_checker_1.doesHaveValue(e)) {
                error = util_1.default.format(e);
            }
            else {
                error = new Error('Promise rejected without a reason');
            }
        }
        time_1.default.clearTimeout(timeoutId);
        uncaught_exception_manager_1.default.unregisterHandler(exceptionHandler);
        return { error, result };
    },
};
exports.default = UserCodeRunner;
//# sourceMappingURL=user_code_runner.js.map