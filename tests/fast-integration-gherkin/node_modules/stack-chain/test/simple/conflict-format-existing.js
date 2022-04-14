
var test = require("tap").test;
var defaultFormater = require('../../format.js');
var produce = require('../produce.js');

// Set a formater before stack-chain is required
Error.prepareStackTrace = function (error, frames) {
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

var chain = require('../../');

test("set Error.prepareStackTrace before require", function (t) {
  t.test("default formatter replaced", function (t) {
    t.equal(produce.real(3), produce.fake([
      'Error: trace',
      '',
      'deepStack',
      'deepStack'
    ]));

    t.end();
  });

  t.test("restore default formater", function (t) {
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
