"use strict";
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    Object.defineProperty(o, k2, { enumerable: true, get: function() { return m[k]; } });
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || function (mod) {
    if (mod && mod.__esModule) return mod;
    var result = {};
    if (mod != null) for (var k in mod) if (k !== "default" && Object.prototype.hasOwnProperty.call(mod, k)) __createBinding(result, mod, k);
    __setModuleDefault(result, mod);
    return result;
};
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const TreeRegexp_1 = __importDefault(require("../src/TreeRegexp"));
const ParameterTypeRegistry_1 = __importDefault(require("../src/ParameterTypeRegistry"));
const Argument_1 = __importDefault(require("../src/Argument"));
const assert = __importStar(require("assert"));
describe('Argument', () => {
    it('exposes getParameterTypeName()', () => {
        const treeRegexp = new TreeRegexp_1.default('three (.*) mice');
        const parameterTypeRegistry = new ParameterTypeRegistry_1.default();
        const args = Argument_1.default.build(treeRegexp, 'three blind mice', [
            parameterTypeRegistry.lookupByTypeName('string'),
        ]);
        const argument = args[0];
        assert.strictEqual(argument.getParameterType().name, 'string');
    });
});
//# sourceMappingURL=ArgumentTest.js.map