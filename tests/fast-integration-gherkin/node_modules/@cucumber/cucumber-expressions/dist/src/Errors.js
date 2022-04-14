"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.createUndefinedParameterType = exports.UndefinedParameterTypeError = exports.AmbiguousParameterTypeError = exports.createInvalidParameterTypeNameInNode = exports.createCantEscaped = exports.createAlternationNotAllowedInOptional = exports.createMissingEndToken = exports.createTheEndOfLIneCanNotBeEscaped = exports.createOptionalIsNotAllowedInOptional = exports.createParameterIsNotAllowedInOptional = exports.createOptionalMayNotBeEmpty = exports.createAlternativeMayNotBeEmpty = exports.createAlternativeMayNotExclusivelyContainOptionals = void 0;
const Ast_1 = require("./Ast");
const CucumberExpressionError_1 = __importDefault(require("./CucumberExpressionError"));
function createAlternativeMayNotExclusivelyContainOptionals(node, expression) {
    return new CucumberExpressionError_1.default(message(node.start, expression, pointAtLocated(node), 'An alternative may not exclusively contain optionals', "If you did not mean to use an optional you can use '\\(' to escape the the '('"));
}
exports.createAlternativeMayNotExclusivelyContainOptionals = createAlternativeMayNotExclusivelyContainOptionals;
function createAlternativeMayNotBeEmpty(node, expression) {
    return new CucumberExpressionError_1.default(message(node.start, expression, pointAtLocated(node), 'Alternative may not be empty', "If you did not mean to use an alternative you can use '\\/' to escape the the '/'"));
}
exports.createAlternativeMayNotBeEmpty = createAlternativeMayNotBeEmpty;
function createOptionalMayNotBeEmpty(node, expression) {
    return new CucumberExpressionError_1.default(message(node.start, expression, pointAtLocated(node), 'An optional must contain some text', "If you did not mean to use an optional you can use '\\(' to escape the the '('"));
}
exports.createOptionalMayNotBeEmpty = createOptionalMayNotBeEmpty;
function createParameterIsNotAllowedInOptional(node, expression) {
    return new CucumberExpressionError_1.default(message(node.start, expression, pointAtLocated(node), 'An optional may not contain a parameter type', "If you did not mean to use an parameter type you can use '\\{' to escape the the '{'"));
}
exports.createParameterIsNotAllowedInOptional = createParameterIsNotAllowedInOptional;
function createOptionalIsNotAllowedInOptional(node, expression) {
    return new CucumberExpressionError_1.default(message(node.start, expression, pointAtLocated(node), 'An optional may not contain an other optional', "If you did not mean to use an optional type you can use '\\(' to escape the the '('. For more complicated expressions consider using a regular expression instead."));
}
exports.createOptionalIsNotAllowedInOptional = createOptionalIsNotAllowedInOptional;
function createTheEndOfLIneCanNotBeEscaped(expression) {
    const index = Array.from(expression).length - 1;
    return new CucumberExpressionError_1.default(message(index, expression, pointAt(index), 'The end of line can not be escaped', "You can use '\\\\' to escape the the '\\'"));
}
exports.createTheEndOfLIneCanNotBeEscaped = createTheEndOfLIneCanNotBeEscaped;
function createMissingEndToken(expression, beginToken, endToken, current) {
    const beginSymbol = (0, Ast_1.symbolOf)(beginToken);
    const endSymbol = (0, Ast_1.symbolOf)(endToken);
    const purpose = (0, Ast_1.purposeOf)(beginToken);
    return new CucumberExpressionError_1.default(message(current.start, expression, pointAtLocated(current), `The '${beginSymbol}' does not have a matching '${endSymbol}'`, `If you did not intend to use ${purpose} you can use '\\${beginSymbol}' to escape the ${purpose}`));
}
exports.createMissingEndToken = createMissingEndToken;
function createAlternationNotAllowedInOptional(expression, current) {
    return new CucumberExpressionError_1.default(message(current.start, expression, pointAtLocated(current), 'An alternation can not be used inside an optional', "You can use '\\/' to escape the the '/'"));
}
exports.createAlternationNotAllowedInOptional = createAlternationNotAllowedInOptional;
function createCantEscaped(expression, index) {
    return new CucumberExpressionError_1.default(message(index, expression, pointAt(index), "Only the characters '{', '}', '(', ')', '\\', '/' and whitespace can be escaped", "If you did mean to use an '\\' you can use '\\\\' to escape it"));
}
exports.createCantEscaped = createCantEscaped;
function createInvalidParameterTypeNameInNode(token, expression) {
    return new CucumberExpressionError_1.default(message(token.start, expression, pointAtLocated(token), "Parameter names may not contain '{', '}', '(', ')', '\\' or '/'", 'Did you mean to use a regular expression?'));
}
exports.createInvalidParameterTypeNameInNode = createInvalidParameterTypeNameInNode;
function message(index, expression, pointer, problem, solution) {
    return `This Cucumber Expression has a problem at column ${index + 1}:

${expression}
${pointer}
${problem}.
${solution}`;
}
function pointAt(index) {
    const pointer = [];
    for (let i = 0; i < index; i++) {
        pointer.push(' ');
    }
    pointer.push('^');
    return pointer.join('');
}
function pointAtLocated(node) {
    const pointer = [pointAt(node.start)];
    if (node.start + 1 < node.end) {
        for (let i = node.start + 1; i < node.end - 1; i++) {
            pointer.push('-');
        }
        pointer.push('^');
    }
    return pointer.join('');
}
class AmbiguousParameterTypeError extends CucumberExpressionError_1.default {
    static forConstructor(keyName, keyValue, parameterTypes, generatedExpressions) {
        return new this(`parameter type with ${keyName}=${keyValue} is used by several parameter types: ${parameterTypes}, ${generatedExpressions}`);
    }
    static forRegExp(parameterTypeRegexp, expressionRegexp, parameterTypes, generatedExpressions) {
        return new this(`Your Regular Expression ${expressionRegexp}
matches multiple parameter types with regexp ${parameterTypeRegexp}:
   ${this._parameterTypeNames(parameterTypes)}

I couldn't decide which one to use. You have two options:

1) Use a Cucumber Expression instead of a Regular Expression. Try one of these:
   ${this._expressions(generatedExpressions)}

2) Make one of the parameter types preferential and continue to use a Regular Expression.
`);
    }
    static _parameterTypeNames(parameterTypes) {
        return parameterTypes.map((p) => `{${p.name}}`).join('\n   ');
    }
    static _expressions(generatedExpressions) {
        return generatedExpressions.map((e) => e.source).join('\n   ');
    }
}
exports.AmbiguousParameterTypeError = AmbiguousParameterTypeError;
class UndefinedParameterTypeError extends CucumberExpressionError_1.default {
    constructor(undefinedParameterTypeName, message) {
        super(message);
        this.undefinedParameterTypeName = undefinedParameterTypeName;
    }
}
exports.UndefinedParameterTypeError = UndefinedParameterTypeError;
function createUndefinedParameterType(node, expression, undefinedParameterTypeName) {
    return new UndefinedParameterTypeError(undefinedParameterTypeName, message(node.start, expression, pointAtLocated(node), `Undefined parameter type '${undefinedParameterTypeName}'`, `Please register a ParameterType for '${undefinedParameterTypeName}'`));
}
exports.createUndefinedParameterType = createUndefinedParameterType;
//# sourceMappingURL=Errors.js.map