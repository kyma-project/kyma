var test = require("tap").test;
var chain = require('../../');

test("non extensible Error objects don't throw", function(t) {
  var error = new Error("don't extend me");
  Object.preventExtensions(error)
  t.doesNotThrow(function() {
    error.stack;
  });
  t.end();
});

test('stack is correct on non extensible error object', function (t) {
  var error = new Error("don't extend me");
  Object.preventExtensions(error);

  chain.format.replace(function () {
    return 'good';
  });

  try {
    t.equal(error.stack, 'good');
  } catch (e) { t.ifError(e); }

  chain.format.restore();

  t.end();
});

