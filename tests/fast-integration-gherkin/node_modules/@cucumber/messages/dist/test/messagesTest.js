"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const assert_1 = __importDefault(require("assert"));
const index_1 = require("../src/index");
describe('messages', () => {
    it('defaults missing fields when deserialising from JSON', () => {
        // Sample envelope from before we moved from protobuf to JSON Schema
        const partialGherkinDocumentEnvelope = {
            gherkinDocument: {
                feature: {
                    children: [
                        {
                            scenario: {
                                id: '1',
                                keyword: 'Scenario',
                                location: { column: 3, line: 3 },
                                name: 'minimalistic',
                                steps: [
                                    {
                                        id: '0',
                                        keyword: 'Given ',
                                        location: { column: 5, line: 4 },
                                        text: 'the minimalism',
                                    },
                                ],
                            },
                        },
                    ],
                    keyword: 'Feature',
                    language: 'en',
                    location: { column: 1, line: 1 },
                    name: 'Minimal',
                },
                uri: 'testdata/good/minimal.feature',
            },
        };
        const envelope = index_1.parseEnvelope(JSON.stringify(partialGherkinDocumentEnvelope));
        const expectedEnvelope = {
            gherkinDocument: {
                // new
                comments: [],
                feature: {
                    // new
                    tags: [],
                    // new
                    description: '',
                    children: [
                        {
                            scenario: {
                                // new
                                examples: [],
                                // new
                                description: '',
                                // new
                                tags: [],
                                id: '1',
                                keyword: 'Scenario',
                                location: { column: 3, line: 3 },
                                name: 'minimalistic',
                                steps: [
                                    {
                                        id: '0',
                                        keyword: 'Given ',
                                        location: { column: 5, line: 4 },
                                        text: 'the minimalism',
                                    },
                                ],
                            },
                        },
                    ],
                    keyword: 'Feature',
                    language: 'en',
                    location: { column: 1, line: 1 },
                    name: 'Minimal',
                },
                uri: 'testdata/good/minimal.feature',
            },
        };
        assert_1.default.deepStrictEqual(JSON.parse(JSON.stringify(envelope)), JSON.parse(JSON.stringify(expectedEnvelope)));
    });
});
//# sourceMappingURL=messagesTest.js.map