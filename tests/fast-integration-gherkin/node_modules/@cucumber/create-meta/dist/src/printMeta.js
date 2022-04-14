"use strict";
var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.main = void 0;
const fs_1 = require("fs");
const createMeta_1 = require("./createMeta");
const ciDict_json_1 = __importDefault(require("./ciDict.json"));
function main(envPath, stdout) {
    return __awaiter(this, void 0, void 0, function* () {
        const envData = yield fs_1.promises.readFile(envPath, 'utf-8');
        const entries = envData.split('\n').map((line) => line.split('='));
        const env = Object.fromEntries(entries);
        const ci = createMeta_1.detectCI(ciDict_json_1.default, env);
        stdout.write(JSON.stringify(ci, null, 2) + '\n');
    });
}
exports.main = main;
main(process.argv[2], process.stdout).catch((err) => console.error(err.backtrace));
//# sourceMappingURL=printMeta.js.map