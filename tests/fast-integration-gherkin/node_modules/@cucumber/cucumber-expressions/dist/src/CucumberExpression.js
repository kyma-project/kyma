"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const TreeRegexp_1 = __importDefault(require("./TreeRegexp"));
const Argument_1 = __importDefault(require("./Argument"));
const Errors_1 = require("./Errors");
const CucumberExpressionParser_1 = __importDefault(require("./CucumberExpressionParser"));
const Ast_1 = require("./Ast");
const ESCAPE_PATTERN = () => /([\\^[({$.|?*+})\]])/g;
class CucumberExpression {
    /**
     * @param expression
     * @param parameterTypeRegistry
     */
    constructor(expression, parameterTypeRegistry) {
        this.expression = expression;
        this.parameterTypeRegistry = parameterTypeRegistry;
        this.parameterTypes = [];
        const parser = new CucumberExpressionParser_1.default();
        const ast = parser.parse(expression);
        const pattern = this.rewriteToRegex(ast);
        this.treeRegexp = new TreeRegexp_1.default(pattern);
    }
    rewriteToRegex(node) {
        switch (node.type) {
            case Ast_1.NodeType.text:
                return CucumberExpression.escapeRegex(node.text());
            case Ast_1.NodeType.optional:
                return this.rewriteOptional(node);
            case Ast_1.NodeType.alternation:
                return this.rewriteAlternation(node);
            case Ast_1.NodeType.alternative:
                return this.rewriteAlternative(node);
            case Ast_1.NodeType.parameter:
                return this.rewriteParameter(node);
            case Ast_1.NodeType.expression:
                return this.rewriteExpression(node);
            default:
                // Can't happen as long as the switch case is exhaustive
                throw new Error(node.type);
        }
    }
    static escapeRegex(expression) {
        return expression.replace(ESCAPE_PATTERN(), '\\$1');
    }
    rewriteOptional(node) {
        this.assertNoParameters(node, (astNode) => (0, Errors_1.createParameterIsNotAllowedInOptional)(astNode, this.expression));
        this.assertNoOptionals(node, (astNode) => (0, Errors_1.createOptionalIsNotAllowedInOptional)(astNode, this.expression));
        this.assertNotEmpty(node, (astNode) => (0, Errors_1.createOptionalMayNotBeEmpty)(astNode, this.expression));
        const regex = node.nodes.map((node) => this.rewriteToRegex(node)).join('');
        return `(?:${regex})?`;
    }
    rewriteAlternation(node) {
        // Make sure the alternative parts aren't empty and don't contain parameter types
        node.nodes.forEach((alternative) => {
            if (alternative.nodes.length == 0) {
                throw (0, Errors_1.createAlternativeMayNotBeEmpty)(alternative, this.expression);
            }
            this.assertNotEmpty(alternative, (astNode) => (0, Errors_1.createAlternativeMayNotExclusivelyContainOptionals)(astNode, this.expression));
        });
        const regex = node.nodes.map((node) => this.rewriteToRegex(node)).join('|');
        return `(?:${regex})`;
    }
    rewriteAlternative(node) {
        return node.nodes.map((lastNode) => this.rewriteToRegex(lastNode)).join('');
    }
    rewriteParameter(node) {
        const name = node.text();
        const parameterType = this.parameterTypeRegistry.lookupByTypeName(name);
        if (!parameterType) {
            throw (0, Errors_1.createUndefinedParameterType)(node, this.expression, name);
        }
        this.parameterTypes.push(parameterType);
        const regexps = parameterType.regexpStrings;
        if (regexps.length == 1) {
            return `(${regexps[0]})`;
        }
        return `((?:${regexps.join(')|(?:')}))`;
    }
    rewriteExpression(node) {
        const regex = node.nodes.map((node) => this.rewriteToRegex(node)).join('');
        return `^${regex}$`;
    }
    assertNotEmpty(node, createNodeWasNotEmptyException) {
        const textNodes = node.nodes.filter((astNode) => Ast_1.NodeType.text == astNode.type);
        if (textNodes.length == 0) {
            throw createNodeWasNotEmptyException(node);
        }
    }
    assertNoParameters(node, createNodeContainedAParameterError) {
        const parameterNodes = node.nodes.filter((astNode) => Ast_1.NodeType.parameter == astNode.type);
        if (parameterNodes.length > 0) {
            throw createNodeContainedAParameterError(parameterNodes[0]);
        }
    }
    assertNoOptionals(node, createNodeContainedAnOptionalError) {
        const parameterNodes = node.nodes.filter((astNode) => Ast_1.NodeType.optional == astNode.type);
        if (parameterNodes.length > 0) {
            throw createNodeContainedAnOptionalError(parameterNodes[0]);
        }
    }
    match(text) {
        return Argument_1.default.build(this.treeRegexp, text, this.parameterTypes);
    }
    get regexp() {
        return this.treeRegexp.regexp;
    }
    get source() {
        return this.expression;
    }
}
exports.default = CucumberExpression;
//# sourceMappingURL=CucumberExpression.js.map