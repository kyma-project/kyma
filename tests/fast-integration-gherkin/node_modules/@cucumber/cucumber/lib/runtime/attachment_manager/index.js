"use strict";
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    Object.defineProperty(o, k2, { enumerable: true, get: function() { return m[k]; } });
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || function (mod) {
    if (mod && mod.__esModule) return mod;
    var result = {};
    if (mod != null) for (var k in mod) if (k !== "default" && Object.prototype.hasOwnProperty.call(mod, k)) __createBinding(result, mod, k);
    __setModuleDefault(result, mod);
    return result;
};
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const is_stream_1 = __importDefault(require("is-stream"));
const messages = __importStar(require("@cucumber/messages"));
const value_checker_1 = require("../../value_checker");
class AttachmentManager {
    constructor(onAttachment) {
        this.onAttachment = onAttachment;
    }
    log(text) {
        return this.create(text, 'text/x.cucumber.log+plain');
    }
    create(data, mediaType, callback) {
        if (Buffer.isBuffer(data)) {
            if (value_checker_1.doesNotHaveValue(mediaType)) {
                throw Error('Buffer attachments must specify a media type');
            }
            this.createBufferAttachment(data, mediaType);
        }
        else if (is_stream_1.default.readable(data)) {
            if (value_checker_1.doesNotHaveValue(mediaType)) {
                throw Error('Stream attachments must specify a media type');
            }
            return this.createStreamAttachment(data, mediaType, callback);
        }
        else if (typeof data === 'string') {
            if (value_checker_1.doesNotHaveValue(mediaType)) {
                mediaType = 'text/plain';
            }
            if (mediaType.startsWith('base64:')) {
                this.createStringAttachment(data, {
                    encoding: messages.AttachmentContentEncoding.BASE64,
                    contentType: mediaType.replace('base64:', ''),
                });
            }
            else {
                this.createStringAttachment(data, {
                    encoding: messages.AttachmentContentEncoding.IDENTITY,
                    contentType: mediaType,
                });
            }
        }
        else {
            throw Error('Invalid attachment data: must be a buffer, readable stream, or string');
        }
    }
    createBufferAttachment(data, mediaType) {
        this.createStringAttachment(data.toString('base64'), {
            encoding: messages.AttachmentContentEncoding.BASE64,
            contentType: mediaType,
        });
    }
    createStreamAttachment(data, mediaType, callback) {
        const promise = new Promise((resolve, reject) => {
            const buffers = [];
            data.on('data', (chunk) => {
                buffers.push(chunk);
            });
            data.on('end', () => {
                this.createBufferAttachment(Buffer.concat(buffers), mediaType);
                resolve();
            });
            data.on('error', reject);
        });
        if (value_checker_1.doesHaveValue(callback)) {
            promise.then(callback, callback);
        }
        else {
            return promise;
        }
    }
    createStringAttachment(data, media) {
        this.onAttachment({ data, media });
    }
}
exports.default = AttachmentManager;
//# sourceMappingURL=index.js.map