"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const createMeta_1 = require("../src/createMeta");
const assert_1 = __importDefault(require("assert"));
describe('removeUserInfoFromUrl', () => {
    it('returns undefined for undefined', () => {
        assert_1.default.strictEqual(createMeta_1.removeUserInfoFromUrl(undefined), undefined);
    });
    it('returns null for null', () => {
        assert_1.default.strictEqual(createMeta_1.removeUserInfoFromUrl(null), null);
    });
    it('returns empty string for empty string', () => {
        assert_1.default.strictEqual(createMeta_1.removeUserInfoFromUrl(null), null);
    });
    it('leaves the data intact when no sensitive information is detected', () => {
        assert_1.default.strictEqual(createMeta_1.removeUserInfoFromUrl('pretty safe'), 'pretty safe');
    });
    context('with URLs', () => {
        it('leaves intact when no password is found', () => {
            assert_1.default.strictEqual(createMeta_1.removeUserInfoFromUrl('https://example.com/git/repo.git'), 'https://example.com/git/repo.git');
        });
        it('removes credentials when found', () => {
            assert_1.default.strictEqual(createMeta_1.removeUserInfoFromUrl('http://login@example.com/git/repo.git'), 'http://example.com/git/repo.git');
        });
        it('removes credentials and passwords when found', () => {
            assert_1.default.strictEqual(createMeta_1.removeUserInfoFromUrl('ssh://login:password@example.com/git/repo.git'), 'ssh://example.com/git/repo.git');
        });
    });
});
//# sourceMappingURL=removeUserInfoFromUrlTest.js.map