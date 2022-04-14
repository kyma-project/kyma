"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const Ast_1 = require("./Ast");
const Errors_1 = require("./Errors");
class CucumberExpressionTokenizer {
    tokenize(expression) {
        const codePoints = Array.from(expression);
        const tokens = [];
        let buffer = [];
        let previousTokenType = Ast_1.TokenType.startOfLine;
        let treatAsText = false;
        let escaped = 0;
        let bufferStartIndex = 0;
        function convertBufferToToken(tokenType) {
            let escapeTokens = 0;
            if (tokenType == Ast_1.TokenType.text) {
                escapeTokens = escaped;
                escaped = 0;
            }
            const consumedIndex = bufferStartIndex + buffer.length + escapeTokens;
            const t = new Ast_1.Token(tokenType, buffer.join(''), bufferStartIndex, consumedIndex);
            buffer = [];
            bufferStartIndex = consumedIndex;
            return t;
        }
        function tokenTypeOf(codePoint, treatAsText) {
            if (!treatAsText) {
                return Ast_1.Token.typeOf(codePoint);
            }
            if (Ast_1.Token.canEscape(codePoint)) {
                return Ast_1.TokenType.text;
            }
            throw (0, Errors_1.createCantEscaped)(expression, bufferStartIndex + buffer.length + escaped);
        }
        function shouldCreateNewToken(previousTokenType, currentTokenType) {
            if (currentTokenType != previousTokenType) {
                return true;
            }
            return currentTokenType != Ast_1.TokenType.whiteSpace && currentTokenType != Ast_1.TokenType.text;
        }
        if (codePoints.length == 0) {
            tokens.push(new Ast_1.Token(Ast_1.TokenType.startOfLine, '', 0, 0));
        }
        codePoints.forEach((codePoint) => {
            if (!treatAsText && Ast_1.Token.isEscapeCharacter(codePoint)) {
                escaped++;
                treatAsText = true;
                return;
            }
            const currentTokenType = tokenTypeOf(codePoint, treatAsText);
            treatAsText = false;
            if (shouldCreateNewToken(previousTokenType, currentTokenType)) {
                const token = convertBufferToToken(previousTokenType);
                previousTokenType = currentTokenType;
                buffer.push(codePoint);
                tokens.push(token);
            }
            else {
                previousTokenType = currentTokenType;
                buffer.push(codePoint);
            }
        });
        if (buffer.length > 0) {
            const token = convertBufferToToken(previousTokenType);
            tokens.push(token);
        }
        if (treatAsText) {
            throw (0, Errors_1.createTheEndOfLIneCanNotBeEscaped)(expression);
        }
        tokens.push(new Ast_1.Token(Ast_1.TokenType.endOfLine, '', codePoints.length, codePoints.length));
        return tokens;
    }
}
exports.default = CucumberExpressionTokenizer;
//# sourceMappingURL=CucumberExpressionTokenizer.js.map