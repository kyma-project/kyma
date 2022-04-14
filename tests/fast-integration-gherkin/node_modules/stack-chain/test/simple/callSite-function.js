
var test = require("tap").test;
var chain = require('../../');
var produce = require('../produce.js');

Error.stackTraceLimit = Infinity;

test("stack extend part", function (t) {
  var extend = function (error, frames) {
    frames.splice(1, 0, 'EXTEND', 'FILTER ME');
    return frames;
  };

  var filter = function (error, frames) {
    return frames.filter(function (callSite) {
      return callSite !== 'FILTER ME';
    });
  };

  var callSites = function (level, options) {
    var limit = Error.stackTraceLimit;
    var callSites;
    produce.deepStack(0, level, function () {
      Error.stackTraceLimit = level;
      callSites = chain.callSite(options);
      Error.stackTraceLimit = limit;
    });

    return callSites.slice(1, Infinity);
  };

  t.test("callSite method matches simple case property length", function (t) {
    var method = chain.callSite();
    var propery = chain.originalCallSite(new Error());
    t.strictEqual(method.length, propery.length);

    // The other stuff still works
    t.equal(produce.real(3), produce.fake([
      'Error: trace',
      '    at {where}:18:17',
      '    at deepStack ({where}:5:5)',
      '    at deepStack ({where}:7:5)'
    ]));

    t.end();
  });

  t.test("pretest: toString of callSites array", function (t) {
    t.equal(produce.convert(callSites(3)), produce.fake([
      '    at deepStack ({where}:5:5)',
      '    at deepStack ({where}:7:5)'
    ]));

    t.end();
  });

  t.test("callSite with extend", function (t) {
    chain.extend.attach(extend);
    var textA = produce.convert(callSites(3, { extend: true }));
    var textB = produce.convert(callSites(3));
    chain.extend.deattach(extend);

    t.equal(textA, produce.fake([
      '    at EXTEND',
      '    at FILTER ME',
      '    at deepStack ({where}:5:5)',
      '    at deepStack ({where}:7:5)'
    ]));

    t.equal(textB, produce.fake([
      '    at deepStack ({where}:5:5)',
      '    at deepStack ({where}:7:5)'
    ]));

    t.end();
  });

  t.test("callSite with extend and filter", function (t) {
    chain.extend.attach(extend);
    chain.filter.attach(filter);
    var textA = produce.convert(callSites(3, { extend: true, filter: true }));
    var textB = produce.convert(callSites(3, { filter: true }));
    chain.filter.deattach(filter);
    chain.extend.deattach(extend);

    t.equal(textA, produce.fake([
      '    at EXTEND',
      '    at deepStack ({where}:5:5)',
      '    at deepStack ({where}:7:5)'
    ]));

    t.equal(textB, produce.fake([
      '    at deepStack ({where}:5:5)',
      '    at deepStack ({where}:7:5)'
    ]));

    t.end();
  });

  t.test("callSite with extend and filter and slice", function (t) {
    chain.extend.attach(extend);
    chain.filter.attach(filter);
    var textA = produce.convert(callSites(3, { extend: true, filter: true, slice: 1 }));
    var textB = produce.convert(callSites(3, { slice: 1 }));
    chain.filter.deattach(filter);
    chain.extend.deattach(extend);

    t.equal(textA, produce.fake([
      '    at EXTEND',
      '    at deepStack ({where}:7:5)'
    ]));

    t.equal(textB, produce.fake([
      '    at deepStack ({where}:7:5)'
    ]));

    t.end();
  });

  t.end();
});
