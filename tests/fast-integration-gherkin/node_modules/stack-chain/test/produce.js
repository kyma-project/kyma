
// Produces an error with `level` deept in the call stack
exports.deepStack = function deepStack(curr, top, callback) {
  if (curr === top) {
    callback();
  } else {
    deepStack(curr + 1, top, callback);
  }
};

exports.real = function produceError(level) {
  var stack;
  var limit = Error.stackTraceLimit;

  exports.deepStack(0, level, function () {
    Error.stackTraceLimit = level;

    var error = new Error('trace');
        error.test = true;
    stack = error.stack;

    Error.stackTraceLimit = limit;
  });

  return stack || 'Error: trace';
};

exports.fake = function(input) {
  var output = [];

  for (var i = 0, l = input.length; i < l; i++) {
    output.push(input[i].replace('{where}', module.filename));
  }

  return output.join('\n');
};

exports.convert = function (callSites) {
  var lines = [];
  for (var i = 0; i < callSites.length; i++) {
    lines.push("    at " + callSites[i].toString());
  }
  return lines.join('\n');
};
