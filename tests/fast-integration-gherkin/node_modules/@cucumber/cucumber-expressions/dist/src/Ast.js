"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.TokenType = exports.Token = exports.NodeType = exports.Node = exports.purposeOf = exports.symbolOf = void 0;
const escapeCharacter = '\\';
const alternationCharacter = '/';
const beginParameterCharacter = '{';
const endParameterCharacter = '}';
const beginOptionalCharacter = '(';
const endOptionalCharacter = ')';
function symbolOf(token) {
    switch (token) {
        case TokenType.beginOptional:
            return beginOptionalCharacter;
        case TokenType.endOptional:
            return endOptionalCharacter;
        case TokenType.beginParameter:
            return beginParameterCharacter;
        case TokenType.endParameter:
            return endParameterCharacter;
        case TokenType.alternation:
            return alternationCharacter;
    }
    return '';
}
exports.symbolOf = symbolOf;
function purposeOf(token) {
    switch (token) {
        case TokenType.beginOptional:
        case TokenType.endOptional:
            return 'optional text';
        case TokenType.beginParameter:
        case TokenType.endParameter:
            return 'a parameter';
        case TokenType.alternation:
            return 'alternation';
    }
    return '';
}
exports.purposeOf = purposeOf;
class Node {
    constructor(type, nodes = undefined, token = undefined, start, end) {
        if (nodes === undefined && token === undefined) {
            throw new Error('Either nodes or token must be defined');
        }
        if (nodes === null || token === null) {
            throw new Error('Either nodes or token may not be null');
        }
        this.type = type;
        this.nodes = nodes;
        this.token = token;
        this.start = start;
        this.end = end;
    }
    text() {
        if (this.nodes) {
            return this.nodes.map((value) => value.text()).join('');
        }
        return this.token;
    }
}
exports.Node = Node;
var NodeType;
(function (NodeType) {
    NodeType["text"] = "TEXT_NODE";
    NodeType["optional"] = "OPTIONAL_NODE";
    NodeType["alternation"] = "ALTERNATION_NODE";
    NodeType["alternative"] = "ALTERNATIVE_NODE";
    NodeType["parameter"] = "PARAMETER_NODE";
    NodeType["expression"] = "EXPRESSION_NODE";
})(NodeType = exports.NodeType || (exports.NodeType = {}));
class Token {
    constructor(type, text, start, end) {
        this.type = type;
        this.text = text;
        this.start = start;
        this.end = end;
    }
    static isEscapeCharacter(codePoint) {
        return codePoint == escapeCharacter;
    }
    static canEscape(codePoint) {
        if (codePoint == ' ') {
            // TODO: Unicode whitespace?
            return true;
        }
        switch (codePoint) {
            case escapeCharacter:
                return true;
            case alternationCharacter:
                return true;
            case beginParameterCharacter:
                return true;
            case endParameterCharacter:
                return true;
            case beginOptionalCharacter:
                return true;
            case endOptionalCharacter:
                return true;
        }
        return false;
    }
    static typeOf(codePoint) {
        if (codePoint == ' ') {
            // TODO: Unicode whitespace?
            return TokenType.whiteSpace;
        }
        switch (codePoint) {
            case alternationCharacter:
                return TokenType.alternation;
            case beginParameterCharacter:
                return TokenType.beginParameter;
            case endParameterCharacter:
                return TokenType.endParameter;
            case beginOptionalCharacter:
                return TokenType.beginOptional;
            case endOptionalCharacter:
                return TokenType.endOptional;
        }
        return TokenType.text;
    }
}
exports.Token = Token;
var TokenType;
(function (TokenType) {
    TokenType["startOfLine"] = "START_OF_LINE";
    TokenType["endOfLine"] = "END_OF_LINE";
    TokenType["whiteSpace"] = "WHITE_SPACE";
    TokenType["beginOptional"] = "BEGIN_OPTIONAL";
    TokenType["endOptional"] = "END_OPTIONAL";
    TokenType["beginParameter"] = "BEGIN_PARAMETER";
    TokenType["endParameter"] = "END_PARAMETER";
    TokenType["alternation"] = "ALTERNATION";
    TokenType["text"] = "TEXT";
})(TokenType = exports.TokenType || (exports.TokenType = {}));
//# sourceMappingURL=Ast.js.map