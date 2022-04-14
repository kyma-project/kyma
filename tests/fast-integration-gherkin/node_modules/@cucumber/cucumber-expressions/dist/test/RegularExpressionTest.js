"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const assert_1 = __importDefault(require("assert"));
const RegularExpression_1 = __importDefault(require("../src/RegularExpression"));
const ParameterTypeRegistry_1 = __importDefault(require("../src/ParameterTypeRegistry"));
describe('RegularExpression', () => {
    it('documents match arguments', () => {
        const parameterRegistry = new ParameterTypeRegistry_1.default();
        /// [capture-match-arguments]
        const expr = /I have (\d+) cukes? in my (\w+) now/;
        const expression = new RegularExpression_1.default(expr, parameterRegistry);
        const args = expression.match('I have 7 cukes in my belly now');
        assert_1.default.strictEqual(7, args[0].getValue(null));
        assert_1.default.strictEqual('belly', args[1].getValue(null));
        /// [capture-match-arguments]
    });
    it('does no transform by default', () => {
        assert_1.default.deepStrictEqual(match(/(\d\d)/, '22')[0], '22');
    });
    it('does not transform anonymous', () => {
        assert_1.default.deepStrictEqual(match(/(.*)/, '22')[0], '22');
    });
    it('transforms negative int', () => {
        assert_1.default.deepStrictEqual(match(/(-?\d+)/, '-22')[0], -22);
    });
    it('transforms positive int', () => {
        assert_1.default.deepStrictEqual(match(/(\d+)/, '22')[0], 22);
    });
    it('transforms float without integer part', () => {
        assert_1.default.deepStrictEqual(match(new RegExp(`(${ParameterTypeRegistry_1.default.FLOAT_REGEXP.source})`), '.22')[0], 0.22);
    });
    it('transforms float with sign', () => {
        assert_1.default.deepStrictEqual(match(new RegExp(`(${ParameterTypeRegistry_1.default.FLOAT_REGEXP.source})`), '-1.22')[0], -1.22);
    });
    it('returns null when there is no match', () => {
        assert_1.default.strictEqual(match(/hello/, 'world'), null);
    });
    it('matches nested capture group without match', () => {
        assert_1.default.deepStrictEqual(match(/^a user( named "([^"]*)")?$/, 'a user'), [null]);
    });
    it('matches nested capture group with match', () => {
        assert_1.default.deepStrictEqual(match(/^a user( named "([^"]*)")?$/, 'a user named "Charlie"'), [
            'Charlie',
        ]);
    });
    it('matches capture group nested in optional one', () => {
        const regexp = /^a (pre-commercial transaction |pre buyer fee model )?purchase(?: for \$(\d+))?$/;
        assert_1.default.deepStrictEqual(match(regexp, 'a purchase'), [null, null]);
        assert_1.default.deepStrictEqual(match(regexp, 'a purchase for $33'), [null, 33]);
        assert_1.default.deepStrictEqual(match(regexp, 'a pre buyer fee model purchase'), [
            'pre buyer fee model ',
            null,
        ]);
    });
    it('ignores non capturing groups', () => {
        assert_1.default.deepStrictEqual(match(/(\S+) ?(can|cannot)? (?:delete|cancel) the (\d+)(?:st|nd|rd|th) (attachment|slide) ?(?:upload)?/, 'I can cancel the 1st slide upload'), ['I', 'can', 1, 'slide']);
    });
    it('works with escaped parenthesis', () => {
        assert_1.default.deepStrictEqual(match(/Across the line\(s\)/, 'Across the line(s)'), []);
    });
    it('exposes regexp and source', () => {
        const regexp = /I have (\d+) cukes? in my (.+) now/;
        const expression = new RegularExpression_1.default(regexp, new ParameterTypeRegistry_1.default());
        assert_1.default.deepStrictEqual(expression.regexp, regexp);
        assert_1.default.deepStrictEqual(expression.source, regexp.source);
    });
    it('does not take consider parenthesis in character class as group', function () {
        const expression = new RegularExpression_1.default(/^drawings: ([A-Z_, ()]+)$/, new ParameterTypeRegistry_1.default());
        const args = expression.match('drawings: ONE, TWO(ABC)');
        assert_1.default.strictEqual(args[0].getValue(this), 'ONE, TWO(ABC)');
    });
});
const match = (regexp, text) => {
    const parameterRegistry = new ParameterTypeRegistry_1.default();
    const regularExpression = new RegularExpression_1.default(regexp, parameterRegistry);
    const args = regularExpression.match(text);
    if (!args) {
        return null;
    }
    return args.map((arg) => arg.getValue(null));
};
//# sourceMappingURL=RegularExpressionTest.js.map