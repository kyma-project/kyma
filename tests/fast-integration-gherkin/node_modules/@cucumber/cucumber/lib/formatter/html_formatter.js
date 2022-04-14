"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const _1 = __importDefault(require("."));
const resolve_pkg_1 = __importDefault(require("resolve-pkg"));
const html_formatter_1 = __importDefault(require("@cucumber/html-formatter"));
const value_checker_1 = require("../value_checker");
const stream_1 = require("stream");
const util_1 = require("util");
class HtmlFormatter extends _1.default {
    constructor(options) {
        super(options);
        const cucumberHtmlStream = new html_formatter_1.default(resolve_pkg_1.default('@cucumber/html-formatter', { cwd: __dirname }) +
            '/dist/main.css', resolve_pkg_1.default('@cucumber/html-formatter', { cwd: __dirname }) +
            '/dist/main.js');
        options.eventBroadcaster.on('envelope', (envelope) => {
            cucumberHtmlStream.write(envelope);
            if (value_checker_1.doesHaveValue(envelope.testRunFinished)) {
                cucumberHtmlStream.end();
            }
        });
        cucumberHtmlStream.on('data', (chunk) => this.log(chunk));
        this._finished = util_1.promisify(stream_1.finished)(cucumberHtmlStream);
    }
    async finished() {
        await this._finished;
        await super.finished();
    }
}
exports.default = HtmlFormatter;
//# sourceMappingURL=html_formatter.js.map