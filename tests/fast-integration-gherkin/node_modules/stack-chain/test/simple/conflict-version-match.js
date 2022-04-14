
var test = require("tap").test;

var first = global._stackChain = { version: require('../../package.json').version };
var chain = require('../../');

test("same version but copies", function (t) {
  t.strictEqual(chain, first);
  t.end();
});
