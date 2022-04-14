"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const safe_1 = __importDefault(require("colors/safe"));
safe_1.default.enable();
function getColorFns(enabled) {
    if (enabled) {
        return {
            forStatus(status) {
                return {
                    AMBIGUOUS: safe_1.default.red.bind(safe_1.default),
                    FAILED: safe_1.default.red.bind(safe_1.default),
                    PASSED: safe_1.default.green.bind(safe_1.default),
                    PENDING: safe_1.default.yellow.bind(safe_1.default),
                    SKIPPED: safe_1.default.cyan.bind(safe_1.default),
                    UNDEFINED: safe_1.default.yellow.bind(safe_1.default),
                    UNKNOWN: safe_1.default.yellow.bind(safe_1.default),
                }[status];
            },
            location: safe_1.default.gray.bind(safe_1.default),
            tag: safe_1.default.cyan.bind(safe_1.default),
            diffAdded: safe_1.default.green.bind(safe_1.default),
            diffRemoved: safe_1.default.red.bind(safe_1.default),
            errorMessage: safe_1.default.red.bind(safe_1.default),
            errorStack: safe_1.default.grey.bind(safe_1.default),
        };
    }
    else {
        return {
            forStatus(status) {
                return (x) => x;
            },
            location: (x) => x,
            tag: (x) => x,
            diffAdded: (x) => x,
            diffRemoved: (x) => x,
            errorMessage: (x) => x,
            errorStack: (x) => x,
        };
    }
}
exports.default = getColorFns;
//# sourceMappingURL=get_color_fns.js.map