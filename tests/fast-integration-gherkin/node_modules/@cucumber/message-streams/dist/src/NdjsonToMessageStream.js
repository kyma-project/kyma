"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const stream_1 = require("stream");
const messages_1 = require("@cucumber/messages");
/**
 * Transforms an NDJSON stream to a stream of message objects
 */
class NdjsonToMessageStream extends stream_1.Transform {
    /**
     * Create a new stream
     *
     * @param parseLine a function that parses a line. This function may ignore a line by returning null.
     */
    constructor(parseLine = messages_1.parseEnvelope) {
        super({ writableObjectMode: false, readableObjectMode: true });
        this.parseLine = parseLine;
    }
    _transform(chunk, encoding, callback) {
        if (this.buffer === undefined) {
            this.buffer = '';
        }
        this.buffer += Buffer.isBuffer(chunk) ? chunk.toString('utf-8') : chunk;
        const lines = this.buffer.split('\n');
        this.buffer = lines.pop();
        for (const line of lines) {
            if (line.trim().length > 0) {
                try {
                    const envelope = this.parseLine(line);
                    if (envelope !== null) {
                        this.push(envelope);
                    }
                }
                catch (err) {
                    err.message =
                        err.message +
                            `
Not JSON: '${line}'
`;
                    return callback(err);
                }
            }
        }
        callback();
    }
    _flush(callback) {
        if (this.buffer) {
            try {
                const object = JSON.parse(this.buffer);
                this.push(object);
            }
            catch (err) {
                return callback(new Error(`Not JSONs: ${this.buffer}`));
            }
        }
        callback();
    }
}
exports.default = NdjsonToMessageStream;
//# sourceMappingURL=NdjsonToMessageStream.js.map