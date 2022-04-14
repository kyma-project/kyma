var fs = require('fs')
var marked = require('marked')

var TerminalRenderer = require('marked-terminal')

marked.setOptions({ renderer: new TerminalRenderer() })
console.log(marked(fs.readFileSync(__dirname + '/README.md', 'utf-8')))

module.exports = require('@cucumber/cucumber')
