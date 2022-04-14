
var test = require("tap").test;
var chain = require('../../');
var produce = require('../produce.js');

test("stack extend part", function (t) {
  var modify = function (text) {
    return function (error, frames) {
      if (error.test) {
        frames.splice(1, 0, text);
      }

      return frames;
    };
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
    var wonderLand = modify("wonder land");

    chain.extend.attach(wonderLand);

    t.equal(produce.real(3), produce.fake([
      'Error: trace',
      '    at {where}:18:17',
      '    at wonder land',
      '    at deepStack ({where}:5:5)'
    ]));

    chain.extend.deattach(wonderLand);

    t.end();
  });

  t.test("deattach modifier", function (t) {
    var wonderLand = modify("wonder land");

    chain.extend.attach(wonderLand);
    t.equal(chain.extend.deattach(wonderLand), true);

    t.equal(produce.real(3), produce.fake([
      'Error: trace',
      '    at {where}:18:17',
      '    at deepStack ({where}:5:5)',
      '    at deepStack ({where}:7:5)'
    ]));

    t.equal(chain.extend.deattach(wonderLand), false);

    t.end();
  });

  t.test("execution order", function (t) {
    var wonderLand = modify("wonder land");
    var outerSpace = modify("outer space");

    chain.extend.attach(wonderLand);
    chain.extend.attach(outerSpace);

    t.equal(produce.real(3), produce.fake([
      'Error: trace',
      '    at {where}:18:17',
      '    at outer space',
      '    at wonder land'
    ]));

    chain.extend.deattach(wonderLand);
    chain.extend.deattach(outerSpace);

    t.end();
  });

  t.end();
});
