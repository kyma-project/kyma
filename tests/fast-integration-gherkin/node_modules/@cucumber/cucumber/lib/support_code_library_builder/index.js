"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.SupportCodeLibraryBuilder = exports.builtinParameterTypes = void 0;
const lodash_1 = __importDefault(require("lodash"));
const build_helpers_1 = require("./build_helpers");
const test_case_hook_definition_1 = __importDefault(require("../models/test_case_hook_definition"));
const test_step_hook_definition_1 = __importDefault(require("../models/test_step_hook_definition"));
const test_run_hook_definition_1 = __importDefault(require("../models/test_run_hook_definition"));
const step_definition_1 = __importDefault(require("../models/step_definition"));
const helpers_1 = require("../formatter/helpers");
const validate_arguments_1 = __importDefault(require("./validate_arguments"));
const util_arity_1 = __importDefault(require("util-arity"));
const cucumber_expressions_1 = require("@cucumber/cucumber-expressions");
const value_checker_1 = require("../value_checker");
const finalize_helpers_1 = require("./finalize_helpers");
const world_1 = __importDefault(require("./world"));
exports.builtinParameterTypes = ['int', 'float', 'word', 'string', ''];
class SupportCodeLibraryBuilder {
    constructor() {
        const defineStep = this.defineStep.bind(this);
        this.methods = {
            After: this.defineTestCaseHook(() => this.afterTestCaseHookDefinitionConfigs),
            AfterAll: this.defineTestRunHook(() => this.afterTestRunHookDefinitionConfigs),
            AfterStep: this.defineTestStepHook(() => this.afterTestStepHookDefinitionConfigs),
            Before: this.defineTestCaseHook(() => this.beforeTestCaseHookDefinitionConfigs),
            BeforeAll: this.defineTestRunHook(() => this.beforeTestRunHookDefinitionConfigs),
            BeforeStep: this.defineTestStepHook(() => this.beforeTestStepHookDefinitionConfigs),
            defineParameterType: this.defineParameterType.bind(this),
            defineStep,
            Given: defineStep,
            setDefaultTimeout: (milliseconds) => {
                this.defaultTimeout = milliseconds;
            },
            setDefinitionFunctionWrapper: (fn) => {
                console.log(`setDefinitionFunctionWrapper is deprecated and will be removed in version 8.0.0 of cucumber-js. If this was used to wrap generator functions, please transition to using async / await. If this was used to wrap step definitions, please use BeforeStep / AfterStep hooks instead. If you had other use cases, please create an issue.`);
                this.definitionFunctionWrapper = fn;
            },
            setWorldConstructor: (fn) => {
                this.World = fn;
            },
            Then: defineStep,
            When: defineStep,
        };
    }
    defineParameterType(options) {
        const parameterType = build_helpers_1.buildParameterType(options);
        this.parameterTypeRegistry.defineParameterType(parameterType);
    }
    defineStep(pattern, options, code) {
        if (typeof options === 'function') {
            code = options;
            options = {};
        }
        const { line, uri } = build_helpers_1.getDefinitionLineAndUri(this.cwd);
        validate_arguments_1.default({
            args: { code, pattern, options },
            fnName: 'defineStep',
            location: helpers_1.formatLocation({ line, uri }),
        });
        this.stepDefinitionConfigs.push({
            code,
            line,
            options,
            pattern,
            uri,
        });
    }
    defineTestCaseHook(getCollection) {
        return (options, code) => {
            if (typeof options === 'string') {
                options = { tags: options };
            }
            else if (typeof options === 'function') {
                code = options;
                options = {};
            }
            const { line, uri } = build_helpers_1.getDefinitionLineAndUri(this.cwd);
            validate_arguments_1.default({
                args: { code, options },
                fnName: 'defineTestCaseHook',
                location: helpers_1.formatLocation({ line, uri }),
            });
            getCollection().push({
                code,
                line,
                options,
                uri,
            });
        };
    }
    defineTestStepHook(getCollection) {
        return (options, code) => {
            if (typeof options === 'string') {
                options = { tags: options };
            }
            else if (typeof options === 'function') {
                code = options;
                options = {};
            }
            const { line, uri } = build_helpers_1.getDefinitionLineAndUri(this.cwd);
            validate_arguments_1.default({
                args: { code, options },
                fnName: 'defineTestStepHook',
                location: helpers_1.formatLocation({ line, uri }),
            });
            getCollection().push({
                code,
                line,
                options,
                uri,
            });
        };
    }
    defineTestRunHook(getCollection) {
        return (options, code) => {
            if (typeof options === 'function') {
                code = options;
                options = {};
            }
            const { line, uri } = build_helpers_1.getDefinitionLineAndUri(this.cwd);
            validate_arguments_1.default({
                args: { code, options },
                fnName: 'defineTestRunHook',
                location: helpers_1.formatLocation({ line, uri }),
            });
            getCollection().push({
                code,
                line,
                options,
                uri,
            });
        };
    }
    wrapCode({ code, wrapperOptions, }) {
        if (value_checker_1.doesHaveValue(this.definitionFunctionWrapper)) {
            const codeLength = code.length;
            const wrappedCode = this.definitionFunctionWrapper(code, wrapperOptions);
            if (wrappedCode !== code) {
                return util_arity_1.default(codeLength, wrappedCode);
            }
            return wrappedCode;
        }
        return code;
    }
    buildTestCaseHookDefinitions(configs, canonicalIds) {
        return configs.map(({ code, line, options, uri }, index) => {
            const wrappedCode = this.wrapCode({
                code,
                wrapperOptions: options.wrapperOptions,
            });
            return new test_case_hook_definition_1.default({
                code: wrappedCode,
                id: canonicalIds ? canonicalIds[index] : this.newId(),
                line,
                options,
                unwrappedCode: code,
                uri,
            });
        });
    }
    buildTestStepHookDefinitions(configs) {
        return configs.map(({ code, line, options, uri }) => {
            const wrappedCode = this.wrapCode({
                code,
                wrapperOptions: options.wrapperOptions,
            });
            return new test_step_hook_definition_1.default({
                code: wrappedCode,
                id: this.newId(),
                line,
                options,
                unwrappedCode: code,
                uri,
            });
        });
    }
    buildTestRunHookDefinitions(configs) {
        return configs.map(({ code, line, options, uri }) => {
            const wrappedCode = this.wrapCode({
                code,
                wrapperOptions: options.wrapperOptions,
            });
            return new test_run_hook_definition_1.default({
                code: wrappedCode,
                id: this.newId(),
                line,
                options,
                unwrappedCode: code,
                uri,
            });
        });
    }
    buildStepDefinitions(canonicalIds) {
        const stepDefinitions = [];
        const undefinedParameterTypes = [];
        this.stepDefinitionConfigs.forEach(({ code, line, options, pattern, uri }, index) => {
            let expression;
            if (typeof pattern === 'string') {
                try {
                    expression = new cucumber_expressions_1.CucumberExpression(pattern, this.parameterTypeRegistry);
                }
                catch (e) {
                    if (value_checker_1.doesHaveValue(e.undefinedParameterTypeName)) {
                        undefinedParameterTypes.push({
                            name: e.undefinedParameterTypeName,
                            expression: pattern,
                        });
                        return;
                    }
                    throw e;
                }
            }
            else {
                expression = new cucumber_expressions_1.RegularExpression(pattern, this.parameterTypeRegistry);
            }
            const wrappedCode = this.wrapCode({
                code,
                wrapperOptions: options.wrapperOptions,
            });
            stepDefinitions.push(new step_definition_1.default({
                code: wrappedCode,
                expression,
                id: canonicalIds ? canonicalIds[index] : this.newId(),
                line,
                options,
                pattern,
                unwrappedCode: code,
                uri,
            }));
        });
        return { stepDefinitions, undefinedParameterTypes };
    }
    finalize(canonicalIds) {
        if (value_checker_1.doesNotHaveValue(this.definitionFunctionWrapper)) {
            const definitionConfigs = lodash_1.default.chain([
                this.afterTestCaseHookDefinitionConfigs,
                this.afterTestRunHookDefinitionConfigs,
                this.beforeTestCaseHookDefinitionConfigs,
                this.beforeTestRunHookDefinitionConfigs,
                this.stepDefinitionConfigs,
            ])
                .flatten()
                .value();
            finalize_helpers_1.validateNoGeneratorFunctions({ cwd: this.cwd, definitionConfigs });
        }
        const stepDefinitionsResult = this.buildStepDefinitions(canonicalIds === null || canonicalIds === void 0 ? void 0 : canonicalIds.stepDefinitionIds);
        return {
            afterTestCaseHookDefinitions: this.buildTestCaseHookDefinitions(this.afterTestCaseHookDefinitionConfigs, canonicalIds === null || canonicalIds === void 0 ? void 0 : canonicalIds.afterTestCaseHookDefinitionIds),
            afterTestRunHookDefinitions: this.buildTestRunHookDefinitions(this.afterTestRunHookDefinitionConfigs),
            afterTestStepHookDefinitions: this.buildTestStepHookDefinitions(this.afterTestStepHookDefinitionConfigs),
            beforeTestCaseHookDefinitions: this.buildTestCaseHookDefinitions(this.beforeTestCaseHookDefinitionConfigs, canonicalIds === null || canonicalIds === void 0 ? void 0 : canonicalIds.beforeTestCaseHookDefinitionIds),
            beforeTestRunHookDefinitions: this.buildTestRunHookDefinitions(this.beforeTestRunHookDefinitionConfigs),
            beforeTestStepHookDefinitions: this.buildTestStepHookDefinitions(this.beforeTestStepHookDefinitionConfigs),
            defaultTimeout: this.defaultTimeout,
            parameterTypeRegistry: this.parameterTypeRegistry,
            undefinedParameterTypes: stepDefinitionsResult.undefinedParameterTypes,
            stepDefinitions: stepDefinitionsResult.stepDefinitions,
            World: this.World,
        };
    }
    reset(cwd, newId) {
        this.cwd = cwd;
        this.newId = newId;
        this.afterTestCaseHookDefinitionConfigs = [];
        this.afterTestRunHookDefinitionConfigs = [];
        this.afterTestStepHookDefinitionConfigs = [];
        this.beforeTestCaseHookDefinitionConfigs = [];
        this.beforeTestRunHookDefinitionConfigs = [];
        this.beforeTestStepHookDefinitionConfigs = [];
        this.definitionFunctionWrapper = null;
        this.defaultTimeout = 5000;
        this.parameterTypeRegistry = new cucumber_expressions_1.ParameterTypeRegistry();
        this.stepDefinitionConfigs = [];
        this.World = world_1.default;
    }
}
exports.SupportCodeLibraryBuilder = SupportCodeLibraryBuilder;
exports.default = new SupportCodeLibraryBuilder();
//# sourceMappingURL=index.js.map