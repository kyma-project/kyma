"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const stream_1 = require("stream");
/**
 * Transforms a stream of message objects to NDJSON
 */
class MessageToNdjsonStream extends stream_1.Transform {
    constructor() {
        super({ writableObjectMode: true, readableObjectMode: false });
    }
    _transform(envelope, encoding, callback) {
        const json = JSON.stringify(envelope);
        this.push(json + '\n');
        callback();
    }
}
exports.default = MessageToNdjsonStream;
//# sourceMappingURL=MessageToNdjsonStream.js.map