"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.retriesForPickle = exports.getAmbiguousStepException = void 0;
const location_helpers_1 = require("../formatter/helpers/location_helpers");
const cli_table3_1 = __importDefault(require("cli-table3"));
const indent_string_1 = __importDefault(require("indent-string"));
const pickle_filter_1 = require("../pickle_filter");
function getAmbiguousStepException(stepDefinitions) {
    const table = new cli_table3_1.default({
        chars: {
            bottom: '',
            'bottom-left': '',
            'bottom-mid': '',
            'bottom-right': '',
            left: '',
            'left-mid': '',
            mid: '',
            'mid-mid': '',
            middle: ' - ',
            right: '',
            'right-mid': '',
            top: '',
            'top-left': '',
            'top-mid': '',
            'top-right': '',
        },
        style: {
            border: [],
            'padding-left': 0,
            'padding-right': 0,
        },
    });
    table.push(...stepDefinitions.map((stepDefinition) => {
        const pattern = stepDefinition.pattern.toString();
        return [pattern, location_helpers_1.formatLocation(stepDefinition)];
    }));
    return `${'Multiple step definitions match:' + '\n'}${indent_string_1.default(table.toString(), 2)}`;
}
exports.getAmbiguousStepException = getAmbiguousStepException;
function retriesForPickle(pickle, options) {
    const retries = options.retry;
    if (retries === 0) {
        return 0;
    }
    const retryTagFilter = options.retryTagFilter;
    if (retryTagFilter === '') {
        return retries;
    }
    const pickleTagFilter = new pickle_filter_1.PickleTagFilter(retryTagFilter);
    if (pickleTagFilter.matchesAllTagExpressions(pickle)) {
        return retries;
    }
    return 0;
}
exports.retriesForPickle = retriesForPickle;
//# sourceMappingURL=helpers.js.map