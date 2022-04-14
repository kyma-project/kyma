"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const Parser_1 = require("./Parser");
const gherkin_languages_json_1 = __importDefault(require("./gherkin-languages.json"));
const Errors_1 = require("./Errors");
const DIALECT_DICT = gherkin_languages_json_1.default;
const DEFAULT_DOC_STRING_SEPARATOR = /^(```[`]*)(.*)/;
class GherkinInMarkdownTokenMatcher {
    constructor(defaultDialectName = 'en') {
        this.defaultDialectName = defaultDialectName;
        this.dialect = DIALECT_DICT[defaultDialectName];
        this.nonStarStepKeywords = []
            .concat(this.dialect.given)
            .concat(this.dialect.when)
            .concat(this.dialect.then)
            .concat(this.dialect.and)
            .concat(this.dialect.but)
            .filter((value, index, self) => value !== '* ' && self.indexOf(value) === index);
        this.stepRegexp = new RegExp(`${KeywordPrefix.BULLET}(${this.nonStarStepKeywords.map(escapeRegExp).join('|')})`);
        const headerKeywords = []
            .concat(this.dialect.feature)
            .concat(this.dialect.background)
            .concat(this.dialect.rule)
            .concat(this.dialect.scenarioOutline)
            .concat(this.dialect.scenario)
            .concat(this.dialect.examples)
            .filter((value, index, self) => self.indexOf(value) === index);
        this.headerRegexp = new RegExp(`${KeywordPrefix.HEADER}(${headerKeywords.map(escapeRegExp).join('|')})`);
        this.reset();
    }
    changeDialect(newDialectName, location) {
        const newDialect = DIALECT_DICT[newDialectName];
        if (!newDialect) {
            throw Errors_1.NoSuchLanguageException.create(newDialectName, location);
        }
        this.dialectName = newDialectName;
        this.dialect = newDialect;
    }
    // We've made a deliberate choice not to support `# language: [ISO 639-1]` headers or similar
    // in Markdown. Users should specify a language globally. This can be done in
    // cucumber-js using the --language [ISO 639-1] option.
    match_Language(token) {
        if (!token)
            throw new Error('no token');
        return false;
    }
    match_Empty(token) {
        let result = false;
        if (token.line.isEmpty) {
            result = true;
        }
        if (!this.match_TagLine(token) &&
            !this.match_FeatureLine(token) &&
            !this.match_ScenarioLine(token) &&
            !this.match_BackgroundLine(token) &&
            !this.match_ExamplesLine(token) &&
            !this.match_RuleLine(token) &&
            !this.match_TableRow(token) &&
            !this.match_Comment(token) &&
            !this.match_Language(token) &&
            !this.match_DocStringSeparator(token) &&
            !this.match_EOF(token) &&
            !this.match_StepLine(token)) {
            // neutered
            result = true;
        }
        if (result) {
            token.matchedType = Parser_1.TokenType.Empty;
        }
        return this.setTokenMatched(token, null, result);
    }
    match_Other(token) {
        const text = token.line.getLineText(this.indentToRemove); // take the entire line, except removing DocString indents
        token.matchedType = Parser_1.TokenType.Other;
        token.matchedText = text;
        token.matchedIndent = 0;
        return this.setTokenMatched(token, null, true);
    }
    match_Comment(token) {
        let result = false;
        if (token.line.startsWith('|')) {
            const tableCells = token.line.getTableCells();
            if (this.isGfmTableSeparator(tableCells))
                result = true;
        }
        return this.setTokenMatched(token, null, result);
    }
    match_DocStringSeparator(token) {
        const match = token.line.trimmedLineText.match(this.activeDocStringSeparator);
        const [, newSeparator, mediaType] = match || [];
        let result = false;
        if (newSeparator) {
            if (this.activeDocStringSeparator === DEFAULT_DOC_STRING_SEPARATOR) {
                this.activeDocStringSeparator = new RegExp(`^(${newSeparator})$`);
                this.indentToRemove = 0;
            }
            else {
                this.activeDocStringSeparator = DEFAULT_DOC_STRING_SEPARATOR;
            }
            token.matchedKeyword = newSeparator;
            token.matchedType = Parser_1.TokenType.DocStringSeparator;
            token.matchedText = mediaType || '';
            result = true;
        }
        return this.setTokenMatched(token, null, result);
    }
    match_EOF(token) {
        let result = false;
        if (token.isEof) {
            token.matchedType = Parser_1.TokenType.EOF;
            result = true;
        }
        return this.setTokenMatched(token, null, result);
    }
    match_FeatureLine(token) {
        if (this.matchedFeatureLine) {
            return this.setTokenMatched(token, null, false);
        }
        // We first try to match "# Feature: blah"
        let result = this.matchTitleLine(KeywordPrefix.HEADER, this.dialect.feature, ':', token, Parser_1.TokenType.FeatureLine);
        // If we didn't match "# Feature: blah", we still match this line
        // as a FeatureLine.
        // The reason for this is that users may not want to be constrained by having this as their fist line.
        if (!result) {
            token.matchedType = Parser_1.TokenType.FeatureLine;
            token.matchedText = token.line.trimmedLineText;
            result = this.setTokenMatched(token, null, true);
        }
        this.matchedFeatureLine = result;
        return result;
    }
    match_BackgroundLine(token) {
        return this.matchTitleLine(KeywordPrefix.HEADER, this.dialect.background, ':', token, Parser_1.TokenType.BackgroundLine);
    }
    match_RuleLine(token) {
        return this.matchTitleLine(KeywordPrefix.HEADER, this.dialect.rule, ':', token, Parser_1.TokenType.RuleLine);
    }
    match_ScenarioLine(token) {
        return (this.matchTitleLine(KeywordPrefix.HEADER, this.dialect.scenario, ':', token, Parser_1.TokenType.ScenarioLine) ||
            this.matchTitleLine(KeywordPrefix.HEADER, this.dialect.scenarioOutline, ':', token, Parser_1.TokenType.ScenarioLine));
    }
    match_ExamplesLine(token) {
        return this.matchTitleLine(KeywordPrefix.HEADER, this.dialect.examples, ':', token, Parser_1.TokenType.ExamplesLine);
    }
    match_StepLine(token) {
        return this.matchTitleLine(KeywordPrefix.BULLET, this.nonStarStepKeywords, '', token, Parser_1.TokenType.StepLine);
    }
    matchTitleLine(prefix, keywords, keywordSuffix, token, matchedType) {
        const regexp = new RegExp(`${prefix}(${keywords.map(escapeRegExp).join('|')})${keywordSuffix}(.*)`);
        const match = token.line.match(regexp);
        let indent = token.line.indent;
        let result = false;
        if (match) {
            token.matchedType = matchedType;
            token.matchedKeyword = match[2];
            token.matchedText = match[3].trim();
            indent += match[1].length;
            result = true;
        }
        return this.setTokenMatched(token, indent, result);
    }
    setTokenMatched(token, indent, matched) {
        token.matchedGherkinDialect = this.dialectName;
        token.matchedIndent = indent !== null ? indent : token.line == null ? 0 : token.line.indent;
        token.location.column = token.matchedIndent + 1;
        return matched;
    }
    match_TableRow(token) {
        // Gherkin tables must be indented 2-5 spaces in order to be distinguidedn from non-Gherkin tables
        if (token.line.lineText.match(/^\s\s\s?\s?\s?\|/)) {
            const tableCells = token.line.getTableCells();
            if (this.isGfmTableSeparator(tableCells))
                return false;
            token.matchedKeyword = '|';
            token.matchedType = Parser_1.TokenType.TableRow;
            token.matchedItems = tableCells;
            return true;
        }
        return false;
    }
    isGfmTableSeparator(tableCells) {
        const separatorValues = tableCells
            .map((item) => item.text)
            .filter((value) => value.match(/^:?-+:?$/));
        return separatorValues.length > 0;
    }
    match_TagLine(token) {
        const tags = [];
        let m;
        const re = /`(@[^`]+)`/g;
        do {
            m = re.exec(token.line.trimmedLineText);
            if (m) {
                tags.push({
                    column: token.line.indent + m.index + 2,
                    text: m[1],
                });
            }
        } while (m);
        if (tags.length === 0)
            return false;
        token.matchedType = Parser_1.TokenType.TagLine;
        token.matchedItems = tags;
        return true;
    }
    reset() {
        if (this.dialectName !== this.defaultDialectName) {
            this.changeDialect(this.defaultDialectName);
        }
        this.activeDocStringSeparator = DEFAULT_DOC_STRING_SEPARATOR;
    }
}
exports.default = GherkinInMarkdownTokenMatcher;
var KeywordPrefix;
(function (KeywordPrefix) {
    // https://spec.commonmark.org/0.29/#bullet-list-marker
    KeywordPrefix["BULLET"] = "^(\\s*[*+-]\\s*)";
    KeywordPrefix["HEADER"] = "^(#{1,6}\\s)";
})(KeywordPrefix || (KeywordPrefix = {}));
// https://stackoverflow.com/questions/3115150/how-to-escape-regular-expression-special-characters-using-javascript
function escapeRegExp(text) {
    return text.replace(/[-[\]{}()*+?.,\\^$|#\s]/g, '\\$&');
}
//# sourceMappingURL=GherkinInMarkdownTokenMatcher.js.map