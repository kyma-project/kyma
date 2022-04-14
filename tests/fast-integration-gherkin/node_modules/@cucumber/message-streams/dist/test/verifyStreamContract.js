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
var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const messages = __importStar(require("@cucumber/messages"));
const toArray_1 = __importDefault(require("./toArray"));
const assert = require("assert");
function verifyStreamContract(makeFromMessageStream, makeToMessageStream) {
    describe('contract', () => {
        it('can be serialised over a stream', () => __awaiter(this, void 0, void 0, function* () {
            const fromMessageStream = makeFromMessageStream();
            const toMessageStream = makeToMessageStream();
            fromMessageStream.pipe(toMessageStream);
            const outgoingMessages = [
                {
                    source: {
                        data: 'Feature: Hello',
                        uri: 'hello.feature',
                        mediaType: messages.SourceMediaType.TEXT_X_CUCUMBER_GHERKIN_PLAIN,
                    },
                },
                {
                    attachment: {
                        body: 'hello',
                        contentEncoding: messages.AttachmentContentEncoding.IDENTITY,
                        mediaType: 'text/plain',
                    },
                },
            ];
            for (const outgoingMessage of outgoingMessages) {
                fromMessageStream.write(outgoingMessage);
            }
            fromMessageStream.end();
            const incomingMessages = yield toArray_1.default(toMessageStream);
            assert.deepStrictEqual(JSON.parse(JSON.stringify(incomingMessages)), JSON.parse(JSON.stringify(outgoingMessages)));
        }));
    });
}
exports.default = verifyStreamContract;
//# sourceMappingURL=verifyStreamContract.js.map