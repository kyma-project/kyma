"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.validateInstall = void 0;
const fs_1 = __importDefault(require("mz/fs"));
const path_1 = __importDefault(require("path"));
const resolve_1 = __importDefault(require("resolve"));
const util_1 = require("util");
async function validateInstall(cwd) {
    const projectPath = path_1.default.join(__dirname, '..', '..');
    if (projectPath === cwd) {
        return; // cucumber testing itself
    }
    const currentCucumberPath = require.resolve(projectPath);
    let localCucumberPath;
    try {
        localCucumberPath = await util_1.promisify(resolve_1.default)('@cucumber/cucumber', {
            basedir: cwd,
        });
    }
    catch (e) {
        throw new Error('`@cucumber/cucumber` module not resolvable. Must be locally installed.');
    }
    localCucumberPath = await fs_1.default.realpath(localCucumberPath);
    if (localCucumberPath !== currentCucumberPath) {
        throw new Error(`
      You appear to be executing an install of cucumber (most likely a global install)
      that is different from your local install (the one required in your support files).
      For cucumber to work, you need to execute the same install that is required in your support files.
      Please execute the locally installed version to run your tests.

      Executed Path: ${currentCucumberPath}
      Local Path:    ${localCucumberPath}
      `);
    }
}
exports.validateInstall = validateInstall;
//# sourceMappingURL=install_validator.js.map