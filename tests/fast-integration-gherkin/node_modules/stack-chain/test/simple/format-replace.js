
var test = require("tap").test;
var chain = require('../../');
var defaultFormater = require('../../format.js');
var produce = require('../produce.js');

test("stack format part", function (t) {
  var format = function (error, frames) {
    if (error.test) {
      var lines = [];
          lines.push(error.toString());

      for (var i = 0, l = frames.length; i < l; i++) {
          lines.push(frames[i].getFunctionName());
      }

      return lines.join("\n");
    }

    return defaultFormater(error, frames);
  };

  t.test("no formatter set", function (t) {
    t.equal(produce.real(3), produce.fake([
      'Error: trace',
      '    at {where}:18:17',
      '    at deepStack ({where}:5:5)',
      '    at deepStack ({where}:7:5)'
    ]));

    t.end();
  });

  t.test("default formatter replaced", function (t) {
    chain.format.replace(format);

    t.equal(produce.real(3), produce.fake([
      'Error: trace',
      '',
      'deepStack',
      'deepStack'
    ]));

    chain.format.restore();

    t.end();
  });

  t.test("restore default formater", function (t) {
    chain.format.replace(format);
    chain.format.restore();

    t.equal(produce.real(3), produce.fake([
      'Error: trace',
      '    at {where}:18:17',
      '    at deepStack ({where}:5:5)',
      '    at deepStack ({where}:7:5)'
    ]));

    t.end();
  });

  t.end();
});
