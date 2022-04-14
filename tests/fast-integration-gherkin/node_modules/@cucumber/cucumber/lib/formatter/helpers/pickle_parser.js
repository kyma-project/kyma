"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.getPickleLocation = exports.getPickleStepMap = exports.getStepKeyword = exports.getScenarioDescription = void 0;
const lodash_1 = __importDefault(require("lodash"));
const gherkin_document_parser_1 = require("./gherkin_document_parser");
function getScenarioDescription({ pickle, gherkinScenarioMap, }) {
    return lodash_1.default.chain(pickle.astNodeIds)
        .map((id) => gherkinScenarioMap[id])
        .compact()
        .first()
        .value().description;
}
exports.getScenarioDescription = getScenarioDescription;
function getStepKeyword({ pickleStep, gherkinStepMap, }) {
    return lodash_1.default.chain(pickleStep.astNodeIds)
        .map((id) => gherkinStepMap[id])
        .compact()
        .first()
        .value().keyword;
}
exports.getStepKeyword = getStepKeyword;
function getPickleStepMap(pickle) {
    return lodash_1.default.chain(pickle.steps)
        .map((pickleStep) => [pickleStep.id, pickleStep])
        .fromPairs()
        .value();
}
exports.getPickleStepMap = getPickleStepMap;
function getPickleLocation({ gherkinDocument, pickle, }) {
    const gherkinScenarioLocationMap = gherkin_document_parser_1.getGherkinScenarioLocationMap(gherkinDocument);
    return gherkinScenarioLocationMap[lodash_1.default.last(pickle.astNodeIds)];
}
exports.getPickleLocation = getPickleLocation;
//# sourceMappingURL=pickle_parser.js.map