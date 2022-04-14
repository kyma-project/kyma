"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const snippet_syntax_1 = require("./snippet_syntax");
const CALLBACK_NAME = 'callback';
class JavaScriptSnippetSyntax {
    constructor(snippetInterface) {
        this.snippetInterface = snippetInterface;
    }
    build({ comment, generatedExpressions, functionName, stepParameterNames, }) {
        let functionKeyword = 'function ';
        if (this.snippetInterface === snippet_syntax_1.SnippetInterface.AsyncAwait) {
            functionKeyword = 'async ' + functionKeyword;
        }
        else if (this.snippetInterface === snippet_syntax_1.SnippetInterface.Generator) {
            functionKeyword += '*';
        }
        let implementation;
        if (this.snippetInterface === snippet_syntax_1.SnippetInterface.Callback) {
            implementation = `${CALLBACK_NAME}(null, 'pending');`;
        }
        else {
            implementation = "return 'pending';";
        }
        const definitionChoices = generatedExpressions.map((generatedExpression, index) => {
            const prefix = index === 0 ? '' : '// ';
            const allParameterNames = generatedExpression.parameterNames.concat(stepParameterNames);
            if (this.snippetInterface === snippet_syntax_1.SnippetInterface.Callback) {
                allParameterNames.push(CALLBACK_NAME);
            }
            return `${prefix + functionName}('${generatedExpression.source.replace(/'/g, "\\'")}', ${functionKeyword}(${allParameterNames.join(', ')}) {\n`;
        });
        return (`${definitionChoices.join('')}  // ${comment}\n` +
            `  ${implementation}\n` +
            '});');
    }
}
exports.default = JavaScriptSnippetSyntax;
//# sourceMappingURL=javascript_snippet_syntax.js.map