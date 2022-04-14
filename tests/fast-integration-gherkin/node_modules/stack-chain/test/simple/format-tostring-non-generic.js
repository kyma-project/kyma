
var test = require("tap").test;
var chain = require('../../');
var produce = require('../produce.js');

// See issue https://github.com/AndreasMadsen/stack-chain/issues/12 for
// a detailed explaination.

test("formatter works for non-generic (non-safe) toString", function (t) {
  var base = function () {}
  base.toString = base.toString // sets base.toString to base[[proto]].toString
  Object.setPrototypeOf(base, {}); // sets base[[proto]] = {}

  var error = Object.create(base); // wrap base using prototype chain
  Error.captureStackTrace(error); // prepear error.stack

  t.equal(error.stack.split('\n').length, 11);
  t.end();
});
