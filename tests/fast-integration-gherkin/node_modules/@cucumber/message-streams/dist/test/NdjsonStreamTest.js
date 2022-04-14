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
const src_1 = require("../src");
const assert_1 = __importDefault(require("assert"));
const NdjsonToMessageStream_1 = __importDefault(require("../src/NdjsonToMessageStream"));
const verifyStreamContract_1 = __importDefault(require("./verifyStreamContract"));
const toArray_1 = __importDefault(require("./toArray"));
const messages_1 = require("@cucumber/messages");
describe('NdjsonStream', () => {
    const makeToMessageStream = () => new NdjsonToMessageStream_1.default();
    const makeFromMessageStream = () => new src_1.MessageToNdjsonStream();
    verifyStreamContract_1.default(makeFromMessageStream, makeToMessageStream);
    it('converts a buffer stream written byte by byte', (cb) => {
        const stream = makeToMessageStream();
        const envelope = {
            testStepFinished: {
                testStepResult: {
                    status: messages.TestStepResultStatus.UNKNOWN,
                    duration: { nanos: 0, seconds: 0 },
                    willBeRetried: false,
                },
                testCaseStartedId: '1',
                testStepId: '2',
                timestamp: {
                    seconds: 0,
                    nanos: 0,
                },
            },
        };
        const json = JSON.stringify(envelope);
        stream.on('error', cb);
        stream.on('data', (receivedEnvelope) => {
            assert_1.default.deepStrictEqual(envelope, receivedEnvelope);
            cb();
        });
        const buffer = Buffer.from(json);
        for (let i = 0; i < buffer.length; i++) {
            stream.write(buffer.slice(i, i + 1));
        }
        stream.end();
    });
    it('converts messages to JSON with enums as strings', (cb) => {
        const stream = new src_1.MessageToNdjsonStream();
        stream.on('data', (json) => {
            const ob = JSON.parse(json);
            const expected = {
                testStepFinished: {
                    testStepResult: {
                        status: messages.TestStepResultStatus.UNKNOWN,
                        duration: { nanos: 0, seconds: 0 },
                        willBeRetried: false,
                    },
                    testCaseStartedId: '1',
                    testStepId: '2',
                    timestamp: {
                        seconds: 0,
                        nanos: 0,
                    },
                },
            };
            assert_1.default.deepStrictEqual(ob, expected);
            cb();
        });
        const envelope = {
            testStepFinished: {
                testStepResult: {
                    status: messages.TestStepResultStatus.UNKNOWN,
                    duration: { nanos: 0, seconds: 0 },
                    willBeRetried: false,
                },
                testCaseStartedId: '1',
                testStepId: '2',
                timestamp: {
                    seconds: 0,
                    nanos: 0,
                },
            },
        };
        stream.write(envelope);
    });
    it('ignores empty lines', () => __awaiter(void 0, void 0, void 0, function* () {
        const toMessageStream = makeToMessageStream();
        toMessageStream.write('{}\n{}\n\n{}\n');
        toMessageStream.end();
        const incomingMessages = yield toArray_1.default(toMessageStream);
        assert_1.default.deepStrictEqual(incomingMessages, [new messages_1.Envelope(), new messages_1.Envelope(), new messages_1.Envelope()]);
    }));
    it('includes offending line in error message', () => __awaiter(void 0, void 0, void 0, function* () {
        const toMessageStream = makeToMessageStream();
        yield assert_1.default.rejects(() => __awaiter(void 0, void 0, void 0, function* () {
            toMessageStream.write('{}\n  BLA BLA\n\n{}\n');
            toMessageStream.end();
            yield toArray_1.default(toMessageStream);
        }), {
            message: "Unexpected token B in JSON at position 2\nNot JSON: '  BLA BLA'\n",
        });
    }));
});
//# sourceMappingURL=NdjsonStreamTest.js.map