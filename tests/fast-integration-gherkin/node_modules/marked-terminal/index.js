'use strict';

var chalk = require('chalk');
var Table = require('cli-table3');
var cardinal = require('cardinal');
var emoji = require('node-emoji');
const ansiEscapes = require('ansi-escapes');
const supportsHyperlinks = require('supports-hyperlinks');

var TABLE_CELL_SPLIT = '^*||*^';
var TABLE_ROW_WRAP = '*|*|*|*';
var TABLE_ROW_WRAP_REGEXP = new RegExp(escapeRegExp(TABLE_ROW_WRAP), 'g');

var COLON_REPLACER = '*#COLON|*';
var COLON_REPLACER_REGEXP = new RegExp(escapeRegExp(COLON_REPLACER), 'g');

var TAB_ALLOWED_CHARACTERS = ['\t'];

// HARD_RETURN holds a character sequence used to indicate text has a
// hard (no-reflowing) line break.  Previously \r and \r\n were turned
// into \n in marked's lexer- preprocessing step. So \r is safe to use
// to indicate a hard (non-reflowed) return.
var HARD_RETURN = '\r',
  HARD_RETURN_RE = new RegExp(HARD_RETURN),
  HARD_RETURN_GFM_RE = new RegExp(HARD_RETURN + '|<br />');

var defaultOptions = {
  code: chalk.yellow,
  blockquote: chalk.gray.italic,
  html: chalk.gray,
  heading: chalk.green.bold,
  firstHeading: chalk.magenta.underline.bold,
  hr: chalk.reset,
  listitem: chalk.reset,
  list: list,
  table: chalk.reset,
  paragraph: chalk.reset,
  strong: chalk.bold,
  em: chalk.italic,
  codespan: chalk.yellow,
  del: chalk.dim.gray.strikethrough,
  link: chalk.blue,
  href: chalk.blue.underline,
  text: identity,
  unescape: true,
  emoji: true,
  width: 80,
  showSectionPrefix: true,
  reflowText: false,
  tab: 4,
  tableOptions: {}
};

function Renderer(options, highlightOptions) {
  this.o = Object.assign({}, defaultOptions, options);
  this.tab = sanitizeTab(this.o.tab, defaultOptions.tab);
  this.tableSettings = this.o.tableOptions;
  this.emoji = this.o.emoji ? insertEmojis : identity;
  this.unescape = this.o.unescape ? unescapeEntities : identity;
  this.highlightOptions = highlightOptions || {};

  this.transform = compose(undoColon, this.unescape, this.emoji);
}

// Compute length of str not including ANSI escape codes.
// See http://en.wikipedia.org/wiki/ANSI_escape_code#graphics
function textLength(str) {
  return str.replace(/\u001b\[(?:\d{1,3})(?:;\d{1,3})*m/g, '').length;
}

Renderer.prototype.textLength = textLength;

function fixHardReturn(text, reflow) {
  return reflow ? text.replace(HARD_RETURN, /\n/g) : text;
}

Renderer.prototype.text = function(text) {
  return this.o.text(text);
};

Renderer.prototype.code = function(code, lang, escaped) {
  return section(
    indentify(this.tab, highlight(code, lang, this.o, this.highlightOptions))
  );
};

Renderer.prototype.blockquote = function(quote) {
  return section(this.o.blockquote(indentify(this.tab, quote.trim())));
};

Renderer.prototype.html = function(html) {
  return this.o.html(html);
};

Renderer.prototype.heading = function(text, level, raw) {
  text = this.transform(text);

  var prefix = this.o.showSectionPrefix
    ? new Array(level + 1).join('#') + ' '
    : '';
  text = prefix + text;
  if (this.o.reflowText) {
    text = reflowText(text, this.o.width, this.options.gfm);
  }
  return section(
    level === 1 ? this.o.firstHeading(text) : this.o.heading(text)
  );
};

Renderer.prototype.hr = function() {
  return section(this.o.hr(hr('-', this.o.reflowText && this.o.width)));
};

Renderer.prototype.list = function(body, ordered) {
  body = this.o.list(body, ordered, this.tab);
  return section(fixNestedLists(indentLines(this.tab, body), this.tab));
};

Renderer.prototype.listitem = function(text) {
  var transform = compose(this.o.listitem, this.transform);
  var isNested = text.indexOf('\n') !== -1;
  if (isNested) text = text.trim();

  // Use BULLET_POINT as a marker for ordered or unordered list item
  return '\n' + BULLET_POINT + transform(text);
};

Renderer.prototype.checkbox = function(checked) {
  return '[' + (checked ? 'X' : ' ') + '] ';
};

Renderer.prototype.paragraph = function(text) {
  var transform = compose(this.o.paragraph, this.transform);
  text = transform(text);
  if (this.o.reflowText) {
    text = reflowText(text, this.o.width, this.options.gfm);
  }
  return section(text);
};

Renderer.prototype.table = function(header, body) {
  var table = new Table(
    Object.assign(
      {},
      {
        head: generateTableRow(header)[0]
      },
      this.tableSettings
    )
  );

  generateTableRow(body, this.transform).forEach(function(row) {
    table.push(row);
  });
  return section(this.o.table(table.toString()));
};

Renderer.prototype.tablerow = function(content) {
  return TABLE_ROW_WRAP + content + TABLE_ROW_WRAP + '\n';
};

Renderer.prototype.tablecell = function(content, flags) {
  return content + TABLE_CELL_SPLIT;
};

// span level renderer
Renderer.prototype.strong = function(text) {
  return this.o.strong(text);
};

Renderer.prototype.em = function(text) {
  text = fixHardReturn(text, this.o.reflowText);
  return this.o.em(text);
};

Renderer.prototype.codespan = function(text) {
  text = fixHardReturn(text, this.o.reflowText);
  return this.o.codespan(text.replace(/:/g, COLON_REPLACER));
};

Renderer.prototype.br = function() {
  return this.o.reflowText ? HARD_RETURN : '\n';
};

Renderer.prototype.del = function(text) {
  return this.o.del(text);
};

Renderer.prototype.link = function(href, title, text) {
  if (this.options.sanitize) {
    try {
      var prot = decodeURIComponent(unescape(href))
        .replace(/[^\w:]/g, '')
        .toLowerCase();
    } catch (e) {
      return '';
    }
    if (prot.indexOf('javascript:') === 0) {
      return '';
    }
  }

  var hasText = text && text !== href;

  var out = '';

  if (supportsHyperlinks.stdout) {
    let link = '';
    if (text) {
      link = this.o.href(this.emoji(text));
    } else {
      link = this.o.href(href);
    }
    out = ansiEscapes.link(link, href);
  } else {
    if (hasText) out += this.emoji(text) + ' (';
    out += this.o.href(href);
    if (hasText) out += ')';
  }
  return this.o.link(out);
};

Renderer.prototype.image = function(href, title, text) {
  if (typeof this.o.image === 'function') {
    return this.o.image(href, title, text);
  }
  var out = '![' + text;
  if (title) out += ' – ' + title;
  return out + '](' + href + ')\n';
};

module.exports = Renderer;

// Munge \n's and spaces in "text" so that the number of
// characters between \n's is less than or equal to "width".
function reflowText(text, width, gfm) {
  // Hard break was inserted by Renderer.prototype.br or is
  // <br /> when gfm is true
  var splitRe = gfm ? HARD_RETURN_GFM_RE : HARD_RETURN_RE,
    sections = text.split(splitRe),
    reflowed = [];

  sections.forEach(function(section) {
    // Split the section by escape codes so that we can
    // deal with them separately.
    var fragments = section.split(/(\u001b\[(?:\d{1,3})(?:;\d{1,3})*m)/g);
    var column = 0;
    var currentLine = '';
    var lastWasEscapeChar = false;

    while (fragments.length) {
      var fragment = fragments[0];

      if (fragment === '') {
        fragments.splice(0, 1);
        lastWasEscapeChar = false;
        continue;
      }

      // This is an escape code - leave it whole and
      // move to the next fragment.
      if (!textLength(fragment)) {
        currentLine += fragment;
        fragments.splice(0, 1);
        lastWasEscapeChar = true;
        continue;
      }

      var words = fragment.split(/[ \t\n]+/);

      for (var i = 0; i < words.length; i++) {
        var word = words[i];
        var addSpace = column != 0;
        if (lastWasEscapeChar) addSpace = false;

        // If adding the new word overflows the required width
        if (column + word.length + addSpace > width) {
          if (word.length <= width) {
            // If the new word is smaller than the required width
            // just add it at the beginning of a new line
            reflowed.push(currentLine);
            currentLine = word;
            column = word.length;
          } else {
            // If the new word is longer than the required width
            // split this word into smaller parts.
            var w = word.substr(0, width - column - addSpace);
            if (addSpace) currentLine += ' ';
            currentLine += w;
            reflowed.push(currentLine);
            currentLine = '';
            column = 0;

            word = word.substr(w.length);
            while (word.length) {
              var w = word.substr(0, width);

              if (!w.length) break;

              if (w.length < width) {
                currentLine = w;
                column = w.length;
                break;
              } else {
                reflowed.push(w);
                word = word.substr(width);
              }
            }
          }
        } else {
          if (addSpace) {
            currentLine += ' ';
            column++;
          }

          currentLine += word;
          column += word.length;
        }

        lastWasEscapeChar = false;
      }

      fragments.splice(0, 1);
    }

    if (textLength(currentLine)) reflowed.push(currentLine);
  });

  return reflowed.join('\n');
}

function indentLines(indent, text) {
  return text.replace(/(^|\n)(.+)/g, '$1' + indent + '$2');
}

function indentify(indent, text) {
  if (!text) return text;
  return indent + text.split('\n').join('\n' + indent);
}

var BULLET_POINT_REGEX = '\\*';
var NUMBERED_POINT_REGEX = '\\d+\\.';
var POINT_REGEX =
  '(?:' + [BULLET_POINT_REGEX, NUMBERED_POINT_REGEX].join('|') + ')';

// Prevents nested lists from joining their parent list's last line
function fixNestedLists(body, indent) {
  var regex = new RegExp(
    '' +
    '(\\S(?: |  )?)' + // Last char of current point, plus one or two spaces
    // to allow trailing spaces
    '((?:' +
    indent +
    ')+)' + // Indentation of sub point
      '(' +
      POINT_REGEX +
      '(?:.*)+)$',
    'gm'
  ); // Body of subpoint
  return body.replace(regex, '$1\n' + indent + '$2$3');
}

var isPointedLine = function(line, indent) {
  return line.match('^(?:' + indent + ')*' + POINT_REGEX);
};

function toSpaces(str) {
  return ' '.repeat(str.length);
}

var BULLET_POINT = '* ';
function bulletPointLine(indent, line) {
  return isPointedLine(line, indent) ? line : toSpaces(BULLET_POINT) + line;
}

function bulletPointLines(lines, indent) {
  var transform = bulletPointLine.bind(null, indent);
  return lines
    .split('\n')
    .filter(identity)
    .map(transform)
    .join('\n');
}

var numberedPoint = function(n) {
  return n + '. ';
};
function numberedLine(indent, line, num) {
  return isPointedLine(line, indent)
    ? {
        num: num + 1,
        line: line.replace(BULLET_POINT, numberedPoint(num + 1))
      }
    : {
        num: num,
        line: toSpaces(numberedPoint(num)) + line
      };
}

function numberedLines(lines, indent) {
  var transform = numberedLine.bind(null, indent);
  let num = 0;
  return lines
    .split('\n')
    .filter(identity)
    .map(line => {
      const numbered = transform(line, num);
      num = numbered.num;

      return numbered.line;
    })
    .join('\n');
}

function list(body, ordered, indent) {
  body = body.trim();
  body = ordered ? numberedLines(body, indent) : bulletPointLines(body, indent);
  return body;
}

function section(text) {
  return text + '\n\n';
}

function highlight(code, lang, opts, hightlightOpts) {
  if (chalk.level === 0) return code;

  var style = opts.code;

  code = fixHardReturn(code, opts.reflowText);
  if (lang !== 'javascript' && lang !== 'js') {
    return style(code);
  }

  try {
    return cardinal.highlight(code, hightlightOpts);
  } catch (e) {
    return style(code);
  }
}

function insertEmojis(text) {
  return text.replace(/:([A-Za-z0-9_\-\+]+?):/g, function(emojiString) {
    var emojiSign = emoji.get(emojiString);
    if (!emojiSign) return emojiString;
    return emojiSign + ' ';
  });
}

function hr(inputHrStr, length) {
  length = length || process.stdout.columns;
  return new Array(length).join(inputHrStr);
}

function undoColon(str) {
  return str.replace(COLON_REPLACER_REGEXP, ':');
}

function generateTableRow(text, escape) {
  if (!text) return [];
  escape = escape || identity;
  var lines = escape(text).split('\n');

  var data = [];
  lines.forEach(function(line) {
    if (!line) return;
    var parsed = line
      .replace(TABLE_ROW_WRAP_REGEXP, '')
      .split(TABLE_CELL_SPLIT);

    data.push(parsed.splice(0, parsed.length - 1));
  });
  return data;
}

function escapeRegExp(str) {
  return str.replace(/[\-\[\]\/\{\}\(\)\*\+\?\.\\\^\$\|]/g, '\\$&');
}

function unescapeEntities(html) {
  return html
    .replace(/&amp;/g, '&')
    .replace(/&lt;/g, '<')
    .replace(/&gt;/g, '>')
    .replace(/&quot;/g, '"')
    .replace(/&#39;/g, "'");
}

function identity(str) {
  return str;
}

function compose() {
  var funcs = arguments;
  return function() {
    var args = arguments;
    for (var i = funcs.length; i-- > 0; ) {
      args = [funcs[i].apply(this, args)];
    }
    return args[0];
  };
}

function isAllowedTabString(string) {
  return TAB_ALLOWED_CHARACTERS.some(function(char) {
    return string.match('^(' + char + ')+$');
  });
}

function sanitizeTab(tab, fallbackTab) {
  if (typeof tab === 'number') {
    return new Array(tab + 1).join(' ');
  } else if (typeof tab === 'string' && isAllowedTabString(tab)) {
    return tab;
  } else {
    return new Array(fallbackTab + 1).join(' ');
  }
}
