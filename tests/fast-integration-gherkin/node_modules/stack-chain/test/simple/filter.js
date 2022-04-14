
var test = require("tap").test;
var chain = require('../../');
var produce = require('../produce.js');

test("stack filter part", function (t) {
  var filter = function (error, frames) {
    if (error.test) {
      frames.splice(0, 1);
    }

    return frames;
  };

  t.test("no extend modifier attached", function (t) {
    t.equal(produce.real(3), produce.fake([
      'Error: trace',
      '    at {where}:18:17',
      '    at deepStack ({where}:5:5)',
      '    at deepStack ({where}:7:5)'
    ]));

    t.end();
  });

  t.test("attach modifier", function (t) {
    chain.extend.attach(filter);

    t.equal(produce.real(3), produce.fake([
      'Error: trace',
      '    at deepStack ({where}:5:5)',
      '    at deepStack ({where}:7:5)'
    ]));

    chain.extend.deattach(filter);

    t.end();
  });

  t.test("deattach modifier", function (t) {
    chain.extend.attach(filter);
    t.equal(chain.extend.deattach(filter), true);

    t.equal(produce.real(3), produce.fake([
      'Error: trace',
      '    at {where}:18:17',
      '    at deepStack ({where}:5:5)',
      '    at deepStack ({where}:7:5)'
    ]));

    t.equal(chain.extend.deattach(filter), false);

    t.end();
  });

  t.test("execution order", function (t) {
    chain.extend.attach(filter);
    chain.extend.attach(filter);

    t.equal(produce.real(3), produce.fake([
      'Error: trace',
      '    at deepStack ({where}:7:5)'
    ]));

    chain.extend.deattach(filter);
    chain.extend.deattach(filter);

    t.end();
  });

  t.end();
});
