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
exports.removeUserInfoFromUrl = exports.detectCI = void 0;
const os_1 = __importDefault(require("os"));
const url_1 = require("url");
const messages = __importStar(require("@cucumber/messages"));
const ciDict_json_1 = __importDefault(require("./ciDict.json"));
const evaluateVariableExpression_1 = __importDefault(require("./evaluateVariableExpression"));
function createMeta(toolName, toolVersion, envDict, ciDict) {
    if (ciDict === undefined) {
        ciDict = ciDict_json_1.default;
    }
    return {
        protocolVersion: messages.version,
        implementation: {
            name: toolName,
            version: toolVersion,
        },
        cpu: {
            name: os_1.default.arch(),
        },
        os: {
            name: os_1.default.platform(),
            version: os_1.default.release(),
        },
        runtime: {
            name: 'node.js',
            version: process.versions.node,
        },
        ci: detectCI(ciDict, envDict),
    };
}
exports.default = createMeta;
function detectCI(ciDict, envDict) {
    const detected = [];
    for (const [ciName, ciSystem] of Object.entries(ciDict)) {
        const ci = createCi(ciName, ciSystem, envDict);
        if (ci) {
            detected.push(ci);
        }
    }
    if (detected.length !== 1) {
        return undefined;
    }
    if (detected.length > 1) {
        console.error(`@cucumber/create-meta WARNING: Detected more than one CI: ${JSON.stringify(detected, null, 2)}`);
        console.error('Using the first one.');
    }
    return detected[0];
}
exports.detectCI = detectCI;
function removeUserInfoFromUrl(value) {
    if (!value)
        return value;
    const url = url_1.parse(value);
    if (url.auth === null)
        return value;
    url.auth = null;
    return url_1.format(url);
}
exports.removeUserInfoFromUrl = removeUserInfoFromUrl;
function createCi(ciName, ciSystem, envDict) {
    const url = evaluateVariableExpression_1.default(ciSystem.url, envDict);
    if (url === undefined) {
        // The url is what consumers will use as the primary key for a build
        // If this cannot be determined, we return nothing.
        return undefined;
    }
    let branch = evaluateVariableExpression_1.default(ciSystem.git.branch, envDict);
    return {
        url,
        name: ciName,
        git: {
            remote: removeUserInfoFromUrl(evaluateVariableExpression_1.default(ciSystem.git.remote, envDict)),
            revision: evaluateVariableExpression_1.default(ciSystem.git.revision, envDict),
            branch: branch,
            tag: evaluateVariableExpression_1.default(ciSystem.git.tag, envDict),
        },
    };
}
//# sourceMappingURL=createMeta.js.map