# stack-chain [![Build Status](https://secure.travis-ci.org/AndreasMadsen/stack-chain.png)](http://travis-ci.org/AndreasMadsen/stack-chain)

> API for combining call site modifyers

## Installation

```sheel
npm install stack-chain
```
## API documentation

```JavaScript
var chain = require('stack-chain');
```

When the `Error.stack` getter is executed, the `stack-chain` will perform the
following:

1. execute the `modifiers` attached by `chain.extend`.
2. execute the `modifiers` attached by `chain.filter`.
3. execute the `formater` set by `chain.format.replace`.

### chain.extend.attach(modifier)
### chain.filter.attach(modifier)

Will modify the callSite array. Note you shouldn't format the stack trace.

The `modifier` is a function there takes two arguments `error` and `frames`.

* `error` is the `Error` object.
* `frames` is an array of `callSite` objects, see
  [v8 documentation](https://github.com/v8/v8/wiki/Stack-Trace-API)
  for details.

When the `modifier` is done, it should `return` a modified `frames` array.

```JavaScript
chain.filter.attach(function (error, frames) {

    // Filter out traces related to this file
    var rewrite = frames.filter(function (callSite) {
      return callSite.getFileName() !== module.filename;
    });

    return rewrite;
});
```

### chain.extend.deattach(modifier)
### chain.filter.deattach(modifier)

Removes a `modifier` function from the list of `modifiers`.

```JavaScript
var modifier = function () {};

// Attach modifier function
chain.extend.attach(modifier);

// Deattach modifier function
chain.extend.deattach(modifier);
```

### chain.format.replace(formater)

Replaces the default v8 `formater`. The new `formater` takes a two arguments
`error` and `frames`.

* `error` is the `Error` object.
* `callSites` is an array of `callSite` objects, see
  [v8 documentation](https://github.com/v8/v8/wiki/Stack-Trace-API)
  for details.

When the `formater` is done, it should `return` a `string`. The `string` will
what `Error.stack` returns.

```JavaScript
chain.format.replace(function (error, frames) {
  var lines = [];

  lines.push(error.toString());

  for (var i = 0; i < frames.length; i++) {
    lines.push("    at " + frames[i].toString());
  }

  return lines.join("\n");
});
```

### chain.format.restore()

Will restore the default v8 `formater`. Note that dude to the nature of v8
`Error` objects, if one of the getters `Error.stack` or `Error.callSite` has
already executed, the value of `Error.stack` won't change.

### chain.callSite([options])

This will return the unmodified `callSite` array from the current tick. This
is a performance shortcut, as it does not require generating the `.stack`
string. This behaviour is different from the `Error().callSite` properties.

While this is mostly generating `callSite` in hot code, it can be useful to
do some modification on the array. The `options` object, supports the following:

```javascript
options = {
  // (default false) run the extenders on the callSite array.
  extend: true,

  // (default false) run the filters on the callSite array.
  filter: true,

  // (default 0) before running extend or filter methods, slice of some of the
  // end. This can be useful for hiding the place from where you called this
  // function.
  slice: 2
}
```

### chain.originalCallSite(error)

Returns the original `callSite` array.

### chain.mutatedCallSite(error)

Returns the mutated `callSite` array, that is after `extend` and `filter`
is applied. The array will not exceed the `Error.stackTraceLimit`.

### Error.stackTraceLimit

This limites the size of the `callSites` array. The default value is 10, and
can be set to any positive number including `Infinity`. See
[v8 documentation](https://github.com/v8/v8/wiki/Stack-Trace-API)
for details.

## License

**The software is license under "MIT"**

> Copyright (c) 2012 Andreas Madsen
>
> Permission is hereby granted, free of charge, to any person obtaining a copy
> of this software and associated documentation files (the "Software"), to deal
> in the Software without restriction, including without limitation the rights
> to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
> copies of the Software, and to permit persons to whom the Software is
> furnished to do so, subject to the following conditions:
>
> The above copyright notice and this permission notice shall be included in
> all copies or substantial portions of the Software.
>
> THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
> IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
> FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
> AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
> LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
> OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
> THE SOFTWARE.
