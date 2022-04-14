"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.incrementing = exports.uuid = void 0;
const uuid_1 = require("uuid");
function uuid() {
    return () => uuid_1.v4();
}
exports.uuid = uuid;
function incrementing() {
    let next = 0;
    return () => (next++).toString();
}
exports.incrementing = incrementing;
//# sourceMappingURL=IdGenerator.js.map