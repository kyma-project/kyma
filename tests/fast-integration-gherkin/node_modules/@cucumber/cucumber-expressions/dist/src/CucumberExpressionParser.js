"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const Ast_1 = require("./Ast");
const CucumberExpressionTokenizer_1 = __importDefault(require("./CucumberExpressionTokenizer"));
const Errors_1 = require("./Errors");
/*
 * text := whitespace | ')' | '}' | .
 */
function parseText(expression, tokens, current) {
    const token = tokens[current];
    switch (token.type) {
        case Ast_1.TokenType.whiteSpace:
        case Ast_1.TokenType.text:
        case Ast_1.TokenType.endParameter:
        case Ast_1.TokenType.endOptional:
            return {
                consumed: 1,
                ast: [new Ast_1.Node(Ast_1.NodeType.text, undefined, token.text, token.start, token.end)],
            };
        case Ast_1.TokenType.alternation:
            throw (0, Errors_1.createAlternationNotAllowedInOptional)(expression, token);
        case Ast_1.TokenType.startOfLine:
        case Ast_1.TokenType.endOfLine:
        case Ast_1.TokenType.beginOptional:
        case Ast_1.TokenType.beginParameter:
        default:
            // If configured correctly this will never happen
            return { consumed: 0 };
    }
}
/*
 * parameter := '{' + name* + '}'
 */
function parseName(expression, tokens, current) {
    const token = tokens[current];
    switch (token.type) {
        case Ast_1.TokenType.whiteSpace:
        case Ast_1.TokenType.text:
            return {
                consumed: 1,
                ast: [new Ast_1.Node(Ast_1.NodeType.text, undefined, token.text, token.start, token.end)],
            };
        case Ast_1.TokenType.beginOptional:
        case Ast_1.TokenType.endOptional:
        case Ast_1.TokenType.beginParameter:
        case Ast_1.TokenType.endParameter:
        case Ast_1.TokenType.alternation:
            throw (0, Errors_1.createInvalidParameterTypeNameInNode)(token, expression);
        case Ast_1.TokenType.startOfLine:
        case Ast_1.TokenType.endOfLine:
        default:
            // If configured correctly this will never happen
            return { consumed: 0 };
    }
}
/*
 * parameter := '{' + text* + '}'
 */
const parseParameter = parseBetween(Ast_1.NodeType.parameter, Ast_1.TokenType.beginParameter, Ast_1.TokenType.endParameter, [parseName]);
/*
 * optional := '(' + option* + ')'
 * option := optional | parameter | text
 */
const optionalSubParsers = [];
const parseOptional = parseBetween(Ast_1.NodeType.optional, Ast_1.TokenType.beginOptional, Ast_1.TokenType.endOptional, optionalSubParsers);
optionalSubParsers.push(parseOptional, parseParameter, parseText);
/*
 * alternation := alternative* + ( '/' + alternative* )+
 */
function parseAlternativeSeparator(expression, tokens, current) {
    if (!lookingAt(tokens, current, Ast_1.TokenType.alternation)) {
        return { consumed: 0 };
    }
    const token = tokens[current];
    return {
        consumed: 1,
        ast: [new Ast_1.Node(Ast_1.NodeType.alternative, undefined, token.text, token.start, token.end)],
    };
}
const alternativeParsers = [
    parseAlternativeSeparator,
    parseOptional,
    parseParameter,
    parseText,
];
/*
 * alternation := (?<=left-boundary) + alternative* + ( '/' + alternative* )+ + (?=right-boundary)
 * left-boundary := whitespace | } | ^
 * right-boundary := whitespace | { | $
 * alternative: = optional | parameter | text
 */
const parseAlternation = (expression, tokens, current) => {
    const previous = current - 1;
    if (!lookingAtAny(tokens, previous, [
        Ast_1.TokenType.startOfLine,
        Ast_1.TokenType.whiteSpace,
        Ast_1.TokenType.endParameter,
    ])) {
        return { consumed: 0 };
    }
    const result = parseTokensUntil(expression, alternativeParsers, tokens, current, [
        Ast_1.TokenType.whiteSpace,
        Ast_1.TokenType.endOfLine,
        Ast_1.TokenType.beginParameter,
    ]);
    const subCurrent = current + result.consumed;
    if (!result.ast.some((astNode) => astNode.type == Ast_1.NodeType.alternative)) {
        return { consumed: 0 };
    }
    const start = tokens[current].start;
    const end = tokens[subCurrent].start;
    // Does not consume right hand boundary token
    return {
        consumed: result.consumed,
        ast: [
            new Ast_1.Node(Ast_1.NodeType.alternation, splitAlternatives(start, end, result.ast), undefined, start, end),
        ],
    };
};
/*
 * cucumber-expression :=  ( alternation | optional | parameter | text )*
 */
const parseCucumberExpression = parseBetween(Ast_1.NodeType.expression, Ast_1.TokenType.startOfLine, Ast_1.TokenType.endOfLine, [parseAlternation, parseOptional, parseParameter, parseText]);
class CucumberExpressionParser {
    parse(expression) {
        const tokenizer = new CucumberExpressionTokenizer_1.default();
        const tokens = tokenizer.tokenize(expression);
        const result = parseCucumberExpression(expression, tokens, 0);
        return result.ast[0];
    }
}
exports.default = CucumberExpressionParser;
function parseBetween(type, beginToken, endToken, parsers) {
    return (expression, tokens, current) => {
        if (!lookingAt(tokens, current, beginToken)) {
            return { consumed: 0 };
        }
        let subCurrent = current + 1;
        const result = parseTokensUntil(expression, parsers, tokens, subCurrent, [
            endToken,
            Ast_1.TokenType.endOfLine,
        ]);
        subCurrent += result.consumed;
        // endToken not found
        if (!lookingAt(tokens, subCurrent, endToken)) {
            throw (0, Errors_1.createMissingEndToken)(expression, beginToken, endToken, tokens[current]);
        }
        // consumes endToken
        const start = tokens[current].start;
        const end = tokens[subCurrent].end;
        const consumed = subCurrent + 1 - current;
        const ast = [new Ast_1.Node(type, result.ast, undefined, start, end)];
        return { consumed, ast };
    };
}
function parseToken(expression, parsers, tokens, startAt) {
    for (let i = 0; i < parsers.length; i++) {
        const parse = parsers[i];
        const result = parse(expression, tokens, startAt);
        if (result.consumed != 0) {
            return result;
        }
    }
    // If configured correctly this will never happen
    throw new Error('No eligible parsers for ' + tokens);
}
function parseTokensUntil(expression, parsers, tokens, startAt, endTokens) {
    let current = startAt;
    const size = tokens.length;
    const ast = [];
    while (current < size) {
        if (lookingAtAny(tokens, current, endTokens)) {
            break;
        }
        const result = parseToken(expression, parsers, tokens, current);
        if (result.consumed == 0) {
            // If configured correctly this will never happen
            // Keep to avoid infinite loops
            throw new Error('No eligible parsers for ' + tokens);
        }
        current += result.consumed;
        ast.push(...result.ast);
    }
    return { consumed: current - startAt, ast };
}
function lookingAtAny(tokens, at, tokenTypes) {
    return tokenTypes.some((tokenType) => lookingAt(tokens, at, tokenType));
}
function lookingAt(tokens, at, token) {
    if (at < 0) {
        // If configured correctly this will never happen
        // Keep for completeness
        return token == Ast_1.TokenType.startOfLine;
    }
    if (at >= tokens.length) {
        return token == Ast_1.TokenType.endOfLine;
    }
    return tokens[at].type == token;
}
function splitAlternatives(start, end, alternation) {
    const separators = [];
    const alternatives = [];
    let alternative = [];
    alternation.forEach((n) => {
        if (Ast_1.NodeType.alternative == n.type) {
            separators.push(n);
            alternatives.push(alternative);
            alternative = [];
        }
        else {
            alternative.push(n);
        }
    });
    alternatives.push(alternative);
    return createAlternativeNodes(start, end, separators, alternatives);
}
function createAlternativeNodes(start, end, separators, alternatives) {
    const nodes = [];
    for (let i = 0; i < alternatives.length; i++) {
        const n = alternatives[i];
        if (i == 0) {
            const rightSeparator = separators[i];
            nodes.push(new Ast_1.Node(Ast_1.NodeType.alternative, n, undefined, start, rightSeparator.start));
        }
        else if (i == alternatives.length - 1) {
            const leftSeparator = separators[i - 1];
            nodes.push(new Ast_1.Node(Ast_1.NodeType.alternative, n, undefined, leftSeparator.end, end));
        }
        else {
            const leftSeparator = separators[i - 1];
            const rightSeparator = separators[i];
            nodes.push(new Ast_1.Node(Ast_1.NodeType.alternative, n, undefined, leftSeparator.end, rightSeparator.start));
        }
    }
    return nodes;
}
//# sourceMappingURL=CucumberExpressionParser.js.map