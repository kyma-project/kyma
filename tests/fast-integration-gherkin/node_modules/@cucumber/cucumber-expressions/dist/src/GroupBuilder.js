"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const Group_1 = __importDefault(require("./Group"));
class GroupBuilder {
    constructor() {
        this.capturing = true;
        this.groupBuilders = [];
    }
    add(groupBuilder) {
        this.groupBuilders.push(groupBuilder);
    }
    build(match, nextGroupIndex) {
        const groupIndex = nextGroupIndex();
        const children = this.groupBuilders.map((gb) => gb.build(match, nextGroupIndex));
        const value = match[groupIndex] || undefined;
        const index = match.indices[groupIndex];
        const start = index ? index[0] : undefined;
        const end = index ? index[1] : undefined;
        return new Group_1.default(value, start, end, children);
    }
    setNonCapturing() {
        this.capturing = false;
    }
    get children() {
        return this.groupBuilders;
    }
    moveChildrenTo(groupBuilder) {
        this.groupBuilders.forEach((child) => groupBuilder.add(child));
    }
}
exports.default = GroupBuilder;
//# sourceMappingURL=GroupBuilder.js.map