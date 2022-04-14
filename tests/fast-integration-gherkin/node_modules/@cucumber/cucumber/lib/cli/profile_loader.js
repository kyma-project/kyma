"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const lodash_1 = __importDefault(require("lodash"));
const fs_1 = __importDefault(require("mz/fs"));
const path_1 = __importDefault(require("path"));
const string_argv_1 = __importDefault(require("string-argv"));
const value_checker_1 = require("../value_checker");
class ProfileLoader {
    constructor(directory) {
        this.directory = directory;
    }
    async getDefinitions() {
        const definitionsFilePath = path_1.default.join(this.directory, 'cucumber.js');
        const exists = await fs_1.default.exists(definitionsFilePath);
        if (!exists) {
            return {};
        }
        const definitions = require(definitionsFilePath); // eslint-disable-line @typescript-eslint/no-var-requires
        if (typeof definitions !== 'object') {
            throw new Error(`${definitionsFilePath} does not export an object`);
        }
        return definitions;
    }
    async getArgv(profiles) {
        const definitions = await this.getDefinitions();
        if (profiles.length === 0 && value_checker_1.doesHaveValue(definitions.default)) {
            profiles = ['default'];
        }
        const argvs = profiles.map((profile) => {
            if (value_checker_1.doesNotHaveValue(definitions[profile])) {
                throw new Error(`Undefined profile: ${profile}`);
            }
            return string_argv_1.default(definitions[profile]);
        });
        return lodash_1.default.flatten(argvs);
    }
}
exports.default = ProfileLoader;
//# sourceMappingURL=profile_loader.js.map