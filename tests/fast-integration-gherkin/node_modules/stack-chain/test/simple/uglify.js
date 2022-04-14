
var test = require("tap").test;
var uglify = require("uglify-js");
var path = require("path");

test("can be uglified", function (t) {
  var files = ['format.js', 'index.js', 'stack-chain.js'].map(function (filename) {
    return path.resolve(__dirname, '../../' + filename);
  });
  uglify.minify(files);
  t.end()
});
