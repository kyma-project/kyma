"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.isFileNameInCucumber = void 0;
const lodash_1 = __importDefault(require("lodash"));
const stack_chain_1 = __importDefault(require("stack-chain"));
const path_1 = __importDefault(require("path"));
const value_checker_1 = require("./value_checker");
const projectRootPath = path_1.default.join(__dirname, '..');
const projectChildDirs = ['src', 'lib', 'node_modules'];
function isFileNameInCucumber(fileName) {
    return lodash_1.default.some(projectChildDirs, (dir) => lodash_1.default.startsWith(fileName, path_1.default.join(projectRootPath, dir)));
}
exports.isFileNameInCucumber = isFileNameInCucumber;
class StackTraceFilter {
    filter() {
        this.currentFilter = stack_chain_1.default.filter.attach((_err, frames) => {
            if (this.isErrorInCucumber(frames)) {
                return frames;
            }
            const index = lodash_1.default.findIndex(frames, this.isFrameInCucumber.bind(this));
            if (index === -1) {
                return frames;
            }
            return frames.slice(0, index);
        });
    }
    isErrorInCucumber(frames) {
        const filteredFrames = lodash_1.default.reject(frames, this.isFrameInNode.bind(this));
        return (filteredFrames.length > 0 && this.isFrameInCucumber(filteredFrames[0]));
    }
    isFrameInCucumber(frame) {
        const fileName = value_checker_1.valueOrDefault(frame.getFileName(), '');
        return isFileNameInCucumber(fileName);
    }
    isFrameInNode(frame) {
        const fileName = value_checker_1.valueOrDefault(frame.getFileName(), '');
        return !lodash_1.default.includes(fileName, path_1.default.sep);
    }
    unfilter() {
        stack_chain_1.default.filter.deattach(this.currentFilter);
    }
}
exports.default = StackTraceFilter;
//# sourceMappingURL=stack_trace_filter.js.map