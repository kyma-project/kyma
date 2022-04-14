"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const _1 = __importDefault(require("./"));
const verror_1 = __importDefault(require("verror"));
const publish_banner_1 = __importDefault(require("./publish_banner"));
function exitWithError(error) {
    console.error(verror_1.default.fullStack(error)); // eslint-disable-line no-console
    process.exit(1);
}
function displayPublishAdvertisementBanner() {
    console.error(publish_banner_1.default);
}
async function run() {
    const cwd = process.cwd();
    const cli = new _1.default({
        argv: process.argv,
        cwd,
        stdout: process.stdout,
    });
    let result;
    try {
        result = await cli.run();
    }
    catch (error) {
        exitWithError(error);
    }
    const config = await cli.getConfiguration();
    if (!config.publishing && !config.suppressPublishAdvertisement) {
        displayPublishAdvertisementBanner();
    }
    const exitCode = result.success ? 0 : 1;
    if (result.shouldExitImmediately) {
        process.exit(exitCode);
    }
    else {
        process.exitCode = exitCode;
    }
}
exports.default = run;
//# sourceMappingURL=run.js.map