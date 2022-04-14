"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const message_streams_1 = require("@cucumber/message-streams");
const commander_1 = __importDefault(require("commander"));
const package_json_1 = __importDefault(require("../package.json"));
const stream_1 = require("stream");
const CucumberHtmlStream_1 = __importDefault(require("./CucumberHtmlStream"));
commander_1.default.version(package_json_1.default.version);
commander_1.default.parse(process.argv);
const toMessageStream = new message_streams_1.NdjsonToMessageStream();
stream_1.pipeline(process.stdin, toMessageStream, new CucumberHtmlStream_1.default(__dirname + '/../../dist/main.css', __dirname + '/../../dist/main.js'), process.stdout, (err) => {
    if (err) {
        // tslint:disable-next-line:no-console
        console.error(err);
        process.exit(1);
    }
});
//# sourceMappingURL=cli-main.js.map