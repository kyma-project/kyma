"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const OptionSplitter = {
    split(option) {
        const parts = option.split(/([^A-Z]):(?!\\)/);
        const result = parts.reduce((memo, part, i) => {
            if (partNeedsRecombined(i)) {
                memo.push(parts.slice(i, i + 2).join(''));
            }
            return memo;
        }, []);
        if (result.length === 1) {
            result.push('');
        }
        return result;
    },
};
function partNeedsRecombined(i) {
    return i % 2 === 0;
}
exports.default = OptionSplitter;
//# sourceMappingURL=option_splitter.js.map