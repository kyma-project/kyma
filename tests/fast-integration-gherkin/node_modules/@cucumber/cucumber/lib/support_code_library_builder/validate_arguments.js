"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const lodash_1 = __importDefault(require("lodash"));
const value_checker_1 = require("../value_checker");
const optionsValidation = {
    expectedType: 'object or function',
    predicate({ options }) {
        return lodash_1.default.isPlainObject(options);
    },
};
const optionsTimeoutValidation = {
    identifier: '"options.timeout"',
    expectedType: 'integer',
    predicate({ options }) {
        return value_checker_1.doesNotHaveValue(options.timeout) || lodash_1.default.isInteger(options.timeout);
    },
};
const fnValidation = {
    expectedType: 'function',
    predicate({ code }) {
        return lodash_1.default.isFunction(code);
    },
};
const validations = {
    defineTestRunHook: [
        Object.assign({ identifier: 'first argument' }, optionsValidation),
        optionsTimeoutValidation,
        Object.assign({ identifier: 'second argument' }, fnValidation),
    ],
    defineTestCaseHook: [
        Object.assign({ identifier: 'first argument' }, optionsValidation),
        {
            identifier: '"options.tags"',
            expectedType: 'string',
            predicate({ options }) {
                return value_checker_1.doesNotHaveValue(options.tags) || lodash_1.default.isString(options.tags);
            },
        },
        optionsTimeoutValidation,
        Object.assign({ identifier: 'second argument' }, fnValidation),
    ],
    defineTestStepHook: [
        Object.assign({ identifier: 'first argument' }, optionsValidation),
        {
            identifier: '"options.tags"',
            expectedType: 'string',
            predicate({ options }) {
                return value_checker_1.doesNotHaveValue(options.tags) || lodash_1.default.isString(options.tags);
            },
        },
        optionsTimeoutValidation,
        Object.assign({ identifier: 'second argument' }, fnValidation),
    ],
    defineStep: [
        {
            identifier: 'first argument',
            expectedType: 'string or regular expression',
            predicate({ pattern }) {
                return lodash_1.default.isRegExp(pattern) || lodash_1.default.isString(pattern);
            },
        },
        Object.assign({ identifier: 'second argument' }, optionsValidation),
        optionsTimeoutValidation,
        Object.assign({ identifier: 'third argument' }, fnValidation),
    ],
};
function validateArguments({ args, fnName, location, }) {
    validations[fnName].forEach(({ identifier, expectedType, predicate }) => {
        if (!predicate(args)) {
            throw new Error(`${location}: Invalid ${identifier}: should be a ${expectedType}`);
        }
    });
}
exports.default = validateArguments;
//# sourceMappingURL=validate_arguments.js.map