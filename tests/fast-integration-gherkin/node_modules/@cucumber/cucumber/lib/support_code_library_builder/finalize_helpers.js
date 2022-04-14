"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.validateNoGeneratorFunctions = void 0;
const is_generator_1 = __importDefault(require("is-generator"));
const path_1 = __importDefault(require("path"));
function validateNoGeneratorFunctions({ cwd, definitionConfigs, }) {
    const generatorDefinitionConfigs = definitionConfigs.filter((definitionConfig) => is_generator_1.default.fn(definitionConfig.code));
    if (generatorDefinitionConfigs.length > 0) {
        const references = generatorDefinitionConfigs
            .map((definitionConfig) => `${path_1.default.relative(cwd, definitionConfig.uri)}:${definitionConfig.line.toString()}`)
            .join('\n  ');
        const message = `
      The following hook/step definitions use generator functions:

        ${references}

      Use 'this.setDefinitionFunctionWrapper(fn)' to wrap them in a function that returns a promise.
      `;
        throw new Error(message);
    }
}
exports.validateNoGeneratorFunctions = validateNoGeneratorFunctions;
//# sourceMappingURL=finalize_helpers.js.map