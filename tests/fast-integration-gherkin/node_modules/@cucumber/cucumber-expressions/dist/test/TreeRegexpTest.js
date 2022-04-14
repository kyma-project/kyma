"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const assert_1 = __importDefault(require("assert"));
const TreeRegexp_1 = __importDefault(require("../src/TreeRegexp"));
describe('TreeRegexp', () => {
    it('exposes group source', () => {
        const tr = new TreeRegexp_1.default(/(a(?:b)?)(c)/);
        assert_1.default.deepStrictEqual(tr.groupBuilder.children.map((gb) => gb.source), ['a(?:b)?', 'c']);
    });
    it('builds tree', () => {
        const tr = new TreeRegexp_1.default(/(a(?:b)?)(c)/);
        const group = tr.match('ac');
        assert_1.default.strictEqual(group.value, 'ac');
        assert_1.default.strictEqual(group.children[0].value, 'a');
        assert_1.default.deepStrictEqual(group.children[0].children, []);
        assert_1.default.strictEqual(group.children[1].value, 'c');
    });
    it('ignores `?:` as a non-capturing group', () => {
        const tr = new TreeRegexp_1.default(/a(?:b)(c)/);
        const group = tr.match('abc');
        assert_1.default.strictEqual(group.value, 'abc');
        assert_1.default.strictEqual(group.children.length, 1);
    });
    it('ignores `?!` as a non-capturing group', () => {
        const tr = new TreeRegexp_1.default(/a(?!b)(.+)/);
        const group = tr.match('aBc');
        assert_1.default.strictEqual(group.value, 'aBc');
        assert_1.default.strictEqual(group.children.length, 1);
    });
    it('ignores `?=` as a non-capturing group', () => {
        const tr = new TreeRegexp_1.default(/a(?=[b])(.+)/);
        const group = tr.match('abc');
        assert_1.default.strictEqual(group.value, 'abc');
        assert_1.default.strictEqual(group.children.length, 1);
        assert_1.default.strictEqual(group.children[0].value, 'bc');
    });
    it('ignores `?<=` as a non-capturing group', () => {
        const tr = new TreeRegexp_1.default(/a(.+)(?<=c)$/);
        const group = tr.match('abc');
        assert_1.default.strictEqual(group.value, 'abc');
        assert_1.default.strictEqual(group.children.length, 1);
        assert_1.default.strictEqual(group.children[0].value, 'bc');
    });
    it('ignores `?<!` as a non-capturing group', () => {
        const tr = new TreeRegexp_1.default(/a(.+?)(?<!b)$/);
        const group = tr.match('abc');
        assert_1.default.strictEqual(group.value, 'abc');
        assert_1.default.strictEqual(group.children.length, 1);
        assert_1.default.strictEqual(group.children[0].value, 'bc');
    });
    it('matches named capturing group', () => {
        const tr = new TreeRegexp_1.default(/a(?<name>b)c/);
        const group = tr.match('abc');
        assert_1.default.strictEqual(group.value, 'abc');
        assert_1.default.strictEqual(group.children.length, 1);
        assert_1.default.strictEqual(group.children[0].value, 'b');
    });
    it('matches optional group', () => {
        const tr = new TreeRegexp_1.default(/^Something( with an optional argument)?/);
        const group = tr.match('Something');
        assert_1.default.strictEqual(group.children[0].value, undefined);
    });
    it('matches nested groups', () => {
        const tr = new TreeRegexp_1.default(/^A (\d+) thick line from ((\d+),\s*(\d+),\s*(\d+)) to ((\d+),\s*(\d+),\s*(\d+))/);
        const group = tr.match('A 5 thick line from 10,20,30 to 40,50,60');
        assert_1.default.strictEqual(group.children[0].value, '5');
        assert_1.default.strictEqual(group.children[1].value, '10,20,30');
        assert_1.default.strictEqual(group.children[1].children[0].value, '10');
        assert_1.default.strictEqual(group.children[1].children[1].value, '20');
        assert_1.default.strictEqual(group.children[1].children[2].value, '30');
        assert_1.default.strictEqual(group.children[2].value, '40,50,60');
        assert_1.default.strictEqual(group.children[2].children[0].value, '40');
        assert_1.default.strictEqual(group.children[2].children[1].value, '50');
        assert_1.default.strictEqual(group.children[2].children[2].value, '60');
    });
    it('detects multiple non capturing groups', () => {
        const tr = new TreeRegexp_1.default(/(?:a)(:b)(\?c)(d)/);
        const group = tr.match('a:b?cd');
        assert_1.default.strictEqual(group.children.length, 3);
    });
    it('works with escaped backslash', () => {
        const tr = new TreeRegexp_1.default(/foo\\(bar|baz)/);
        const group = tr.match('foo\\bar');
        assert_1.default.strictEqual(group.children.length, 1);
    });
    it('works with escaped slash', () => {
        const tr = new TreeRegexp_1.default(/^I go to '\/(.+)'$/);
        const group = tr.match("I go to '/hello'");
        assert_1.default.strictEqual(group.children.length, 1);
    });
    it('works with digit and word', () => {
        const tr = new TreeRegexp_1.default(/^(\d) (\w+)$/);
        const group = tr.match('2 you');
        assert_1.default.strictEqual(group.children.length, 2);
    });
    it('captures non capturing groups with capturing groups inside', () => {
        const tr = new TreeRegexp_1.default('the stdout(?: from "(.*?)")?');
        const group = tr.match('the stdout');
        assert_1.default.strictEqual(group.value, 'the stdout');
        assert_1.default.strictEqual(group.children[0].value, undefined);
        assert_1.default.strictEqual(group.children.length, 1);
    });
    it('works with case insensitive flag', () => {
        const tr = new TreeRegexp_1.default(/HELLO/i);
        const group = tr.match('hello');
        assert_1.default.strictEqual(group.value, 'hello');
    });
    it('empty capturing group', () => {
        const tr = new TreeRegexp_1.default(/()/);
        const group = tr.match('');
        // TODO: Would expect the empty string here
        assert_1.default.strictEqual(group.value, undefined);
        assert_1.default.strictEqual(group.children.length, 1);
    });
    it('empty look ahead', () => {
        const tr = new TreeRegexp_1.default(/(?<=)/);
        const group = tr.match('');
        // TODO: Would expect the empty string here
        assert_1.default.strictEqual(group.value, undefined);
        assert_1.default.strictEqual(group.children.length, 0);
    });
    it('does not consider parenthesis in character class as group', () => {
        const tr = new TreeRegexp_1.default(/^drawings: ([A-Z, ()]+)$/);
        const group = tr.match('drawings: ONE(TWO)');
        assert_1.default.strictEqual(group.value, 'drawings: ONE(TWO)');
        assert_1.default.strictEqual(group.children.length, 1);
        assert_1.default.strictEqual(group.children[0].value, 'ONE(TWO)');
    });
});
//# sourceMappingURL=TreeRegexpTest.js.map