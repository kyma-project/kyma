
var test = require("tap").test;
var chain = require('../../');

test("no other copy", function (t) {
  t.strictEqual(global._stackChain, chain);
  t.end();
});
