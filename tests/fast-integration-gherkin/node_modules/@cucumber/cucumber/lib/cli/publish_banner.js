"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const safe_1 = __importDefault(require("colors/safe"));
const cli_table3_1 = __importDefault(require("cli-table3"));
const underlineBoldCyan = (x) => safe_1.default.underline(safe_1.default.bold(safe_1.default.cyan(x)));
const formattedReportUrl = underlineBoldCyan('https://reports.cucumber.io');
const formattedEnv = safe_1.default.cyan('CUCUMBER_PUBLISH_ENABLED') + '=' + safe_1.default.cyan('true');
const formattedMoreInfoUrl = underlineBoldCyan('https://cucumber.io/docs/cucumber/environment-variables/');
const text = `\
Share your Cucumber Report with your team at ${formattedReportUrl}

Command line option:    ${safe_1.default.cyan('--publish')}
Environment variable:   ${formattedEnv}

More information at ${formattedMoreInfoUrl}

To disable this message, add this to your ${safe_1.default.bold('./cucumber.js')}: 
${safe_1.default.bold("module.exports = { default: '--publish-quiet' }")}`;
const table = new cli_table3_1.default({
    style: {
        head: [],
        border: ['green'],
    },
});
table.push([text]);
exports.default = table.toString();
//# sourceMappingURL=publish_banner.js.map