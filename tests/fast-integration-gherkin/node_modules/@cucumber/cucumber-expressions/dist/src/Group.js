"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
class Group {
    constructor(value, start, end, children) {
        this.value = value;
        this.start = start;
        this.end = end;
        this.children = children;
    }
    get values() {
        return (this.children.length === 0 ? [this] : this.children).map((g) => g.value);
    }
}
exports.default = Group;
//# sourceMappingURL=Group.js.map