"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const react_1 = require("@cucumber/react");
const gherkin_utils_1 = require("@cucumber/gherkin-utils");
const query_1 = require("@cucumber/query");
const react_2 = __importDefault(require("react"));
const react_dom_1 = __importDefault(require("react-dom"));
const gherkinQuery = new gherkin_utils_1.Query();
const cucumberQuery = new query_1.Query();
const envelopesQuery = new react_1.EnvelopesQuery();
for (const envelope of window.CUCUMBER_MESSAGES) {
    gherkinQuery.update(envelope);
    cucumberQuery.update(envelope);
    envelopesQuery.update(envelope);
}
const app = (react_2.default.createElement(react_1.QueriesWrapper, { gherkinQuery: gherkinQuery, cucumberQuery: cucumberQuery, envelopesQuery: envelopesQuery },
    react_2.default.createElement(react_1.FilteredResults, null)));
react_dom_1.default.render(app, document.getElementById('content'));
//# sourceMappingURL=main.js.map