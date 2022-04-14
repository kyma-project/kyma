
var test = require("tap").test;
var defaultFormater = require('../../format.js');
var produce = require('../produce.js');

var chain = require('../../');

test("set Error.prepareStackTrace uses stack-chain formater", function (t) {
  // Save original formatter
  var restore = Error.prepareStackTrace;

  // Overwrite formatter
  Error.prepareStackTrace = function (error, frames) {
    if (error.test) {
      Object.defineProperty(error, '__some_secret', {
        value: 'you can\'t compare pain.'
      });
    }

    // Maintain .stack format
    return restore(error, frames);
  };

  // Prope the error using custom prepareStackTrace
  var testError = new Error();
  testError.test = true;
  testError.stack;
  t.equal(testError.__some_secret, 'you can\'t compare pain.');

  // Restore
  Error.prepareStackTrace = restore;

  t.equal(produce.real(3), produce.fake([
    'Error: trace',
    '    at {where}:18:17',
    '    at deepStack ({where}:5:5)',
    '    at deepStack ({where}:7:5)'
  ]));

  t.end();
});

test("set Error.prepareStackTrace uses other formater", function (t) {
  // Another module sets up a formater
  Error.prepareStackTrace = function () {
    return 'custom';
  };

  // Save original formatter
  var restore = Error.prepareStackTrace;

  // Overwrite formatter
  Error.prepareStackTrace = function (error, frames) {
    if (error.test) {
      Object.defineProperty(error, '__some_secret', {
        value: 'you can\'t compare pain.'
      });
    }

    // Maintain .stack format
    return restore(error, frames);
  };

  // Prope the error using custom prepareStackTrace
  var testError = new Error();
  testError.test = true;
  testError.stack;
  t.equal(testError.__some_secret, 'you can\'t compare pain.');

  // Restore
  Error.prepareStackTrace = restore;

  t.equal(produce.real(3), 'custom');

  // Perform an actual restore of the formater, to prevent test conflicts
  chain.format.restore();

  t.equal(produce.real(3), produce.fake([
    'Error: trace',
    '    at {where}:18:17',
    '    at deepStack ({where}:5:5)',
    '    at deepStack ({where}:7:5)'
  ]));

  t.end();
});
