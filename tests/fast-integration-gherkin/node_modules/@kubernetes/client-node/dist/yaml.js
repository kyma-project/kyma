"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.dumpYaml = exports.loadAllYaml = exports.loadYaml = void 0;
const tslib_1 = require("tslib");
const yaml = tslib_1.__importStar(require("js-yaml"));
function loadYaml(data, opts) {
    return yaml.load(data, opts);
}
exports.loadYaml = loadYaml;
function loadAllYaml(data, opts) {
    return yaml.loadAll(data, undefined, opts);
}
exports.loadAllYaml = loadAllYaml;
function dumpYaml(object, opts) {
    return yaml.dump(object, opts);
}
exports.dumpYaml = dumpYaml;
//# sourceMappingURL=yaml.js.map