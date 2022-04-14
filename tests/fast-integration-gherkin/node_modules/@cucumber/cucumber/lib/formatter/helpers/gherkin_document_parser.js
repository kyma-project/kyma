"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.getGherkinScenarioLocationMap = exports.getGherkinExampleRuleMap = exports.getGherkinScenarioMap = exports.getGherkinStepMap = void 0;
const lodash_1 = __importDefault(require("lodash"));
const value_checker_1 = require("../../value_checker");
function getGherkinStepMap(gherkinDocument) {
    return lodash_1.default.chain(gherkinDocument.feature.children)
        .map(extractStepContainers)
        .flatten()
        .map('steps')
        .flatten()
        .map((step) => [step.id, step])
        .fromPairs()
        .value();
}
exports.getGherkinStepMap = getGherkinStepMap;
function extractStepContainers(child) {
    if (value_checker_1.doesHaveValue(child.background)) {
        return [child.background];
    }
    else if (value_checker_1.doesHaveValue(child.rule)) {
        return child.rule.children.map((ruleChild) => value_checker_1.doesHaveValue(ruleChild.background)
            ? ruleChild.background
            : ruleChild.scenario);
    }
    return [child.scenario];
}
function getGherkinScenarioMap(gherkinDocument) {
    return lodash_1.default.chain(gherkinDocument.feature.children)
        .map((child) => {
        if (value_checker_1.doesHaveValue(child.rule)) {
            return child.rule.children;
        }
        return [child];
    })
        .flatten()
        .filter('scenario')
        .map('scenario')
        .map((scenario) => [scenario.id, scenario])
        .fromPairs()
        .value();
}
exports.getGherkinScenarioMap = getGherkinScenarioMap;
function getGherkinExampleRuleMap(gherkinDocument) {
    return lodash_1.default.chain(gherkinDocument.feature.children)
        .filter('rule')
        .map('rule')
        .map((rule) => {
        return rule.children
            .filter((child) => value_checker_1.doesHaveValue(child.scenario))
            .map((child) => [child.scenario.id, rule]);
    })
        .flatten()
        .fromPairs()
        .value();
}
exports.getGherkinExampleRuleMap = getGherkinExampleRuleMap;
function getGherkinScenarioLocationMap(gherkinDocument) {
    const locationMap = {};
    const scenarioMap = getGherkinScenarioMap(gherkinDocument);
    lodash_1.default.entries(scenarioMap).forEach(([id, scenario]) => {
        locationMap[id] = scenario.location;
        if (value_checker_1.doesHaveValue(scenario.examples)) {
            lodash_1.default.chain(scenario.examples)
                .map('tableBody')
                .flatten()
                .forEach((tableRow) => {
                locationMap[tableRow.id] = tableRow.location;
            })
                .value();
        }
    });
    return locationMap;
}
exports.getGherkinScenarioLocationMap = getGherkinScenarioLocationMap;
//# sourceMappingURL=gherkin_document_parser.js.map