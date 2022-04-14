"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.parseEnvelope = void 0;
const messages_1 = require("./messages");
const class_transformer_1 = require("class-transformer");
/**
 * Parses JSON into an Envelope object. The difference from JSON.parse
 * is that the resulting objects will have default values (defined in the JSON Schema)
 * for properties that are absent from the JSON.
 */
function parseEnvelope(json) {
    const plain = JSON.parse(json);
    return class_transformer_1.plainToClass(messages_1.Envelope, plain);
}
exports.parseEnvelope = parseEnvelope;
//# sourceMappingURL=parseEnvelope.js.map