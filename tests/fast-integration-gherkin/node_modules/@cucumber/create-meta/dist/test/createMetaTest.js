"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const createMeta_1 = __importDefault(require("../src/createMeta"));
const assert_1 = __importDefault(require("assert"));
const ciDict_json_1 = __importDefault(require("../src/ciDict.json"));
describe('createMeta', () => {
    it('defines the implementation product', () => {
        const meta = createMeta_1.default('someTool', '1.2.3', {}, {});
        assert_1.default.strictEqual(meta.implementation.name, 'someTool');
        assert_1.default.strictEqual(meta.implementation.version, '1.2.3');
    });
    it('detects CircleCI', () => {
        const envDict = {
            CIRCLE_BUILD_URL: 'the-url',
            CIRCLE_REPOSITORY_URL: 'the-remote',
            CIRCLE_BRANCH: 'the-branch',
            CIRCLE_SHA1: 'the-revision',
            CIRCLE_TAG: 'the-tag',
        };
        const meta = createMeta_1.default('someTool', '1.2.3', envDict, ciDict_json_1.default);
        const ci = {
            name: 'CircleCI',
            url: 'the-url',
            git: {
                remote: 'the-remote',
                branch: 'the-branch',
                revision: 'the-revision',
                tag: 'the-tag',
            },
        };
        assert_1.default.deepStrictEqual(meta.ci, ci);
    });
    it('detects GitHub Actions', () => {
        const envDict = {
            GITHUB_SERVER_URL: 'https://github.com',
            GITHUB_REPOSITORY: 'cucumber/cucumber-ruby',
            GITHUB_RUN_ID: '140170388',
            GITHUB_SHA: 'the-revision',
            GITHUB_REF: 'refs/tags/the-tag',
        };
        const meta = createMeta_1.default('someTool', '1.2.3', envDict, ciDict_json_1.default);
        const ci = {
            name: 'GitHub Actions',
            url: 'https://github.com/cucumber/cucumber-ruby/actions/runs/140170388',
            git: {
                remote: 'https://github.com/cucumber/cucumber-ruby.git',
                branch: undefined,
                revision: 'the-revision',
                tag: 'the-tag',
            },
        };
        assert_1.default.deepStrictEqual(meta.ci, ci);
    });
    it('detects GitHub Actions with custom base url', () => {
        const envDict = {
            GITHUB_SERVER_URL: 'https://github.company.com',
            GITHUB_REPOSITORY: 'cucumber/cucumber-ruby',
            GITHUB_RUN_ID: '140170388',
            GITHUB_SHA: 'the-revision',
            GITHUB_REF: 'refs/heads/the-branch',
        };
        const meta = createMeta_1.default('someTool', '1.2.3', envDict, ciDict_json_1.default);
        const ci = {
            name: 'GitHub Actions',
            url: 'https://github.company.com/cucumber/cucumber-ruby/actions/runs/140170388',
            git: {
                remote: 'https://github.company.com/cucumber/cucumber-ruby.git',
                branch: 'the-branch',
                revision: 'the-revision',
                tag: undefined,
            },
        };
        assert_1.default.deepStrictEqual(meta.ci, ci);
    });
    it('post-processes git refs to branch', () => {
        const envDict = {
            BUILD_URI: 'the-url',
            BUILD_REPOSITORY_URI: 'the-remote',
            BUILD_SOURCEBRANCH: 'refs/heads/main',
            BUILD_SOURCEVERSION: 'the-revision',
        };
        const meta = createMeta_1.default('someTool', '1.2.3', envDict, ciDict_json_1.default);
        const ci = {
            name: 'Azure Pipelines',
            url: 'the-url',
            git: {
                remote: 'the-remote',
                branch: 'main',
                revision: 'the-revision',
                tag: undefined,
            },
        };
        assert_1.default.deepStrictEqual(meta.ci, ci);
    });
    it('post-processes git refs to tag', () => {
        const envDict = {
            BUILD_URI: 'the-url',
            BUILD_REPOSITORY_URI: 'the-remote',
            BUILD_SOURCEBRANCH: 'refs/tags/v1.2.3',
            BUILD_SOURCEVERSION: 'the-revision',
        };
        const meta = createMeta_1.default('someTool', '1.2.3', envDict, ciDict_json_1.default);
        const ci = {
            name: 'Azure Pipelines',
            url: 'the-url',
            git: {
                remote: 'the-remote',
                branch: undefined,
                revision: 'the-revision',
                tag: 'v1.2.3',
            },
        };
        assert_1.default.deepStrictEqual(meta.ci, ci);
    });
});
//# sourceMappingURL=createMetaTest.js.map