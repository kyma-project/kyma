# rfc4648.js

[![Build Status](https://travis-ci.com/swansontec/rfc4648.js.svg?branch=master)](https://travis-ci.com/swansontec/rfc4648.js)
[![JavaScript Style Guide](https://img.shields.io/badge/code_style-standard-brightgreen.svg)](https://standardjs.com)

This library implements encoding and decoding for the data formats specified in [rfc4648](https://tools.ietf.org/html/rfc4648):

- base64
- base64url
- base32
- base32hex
- base16

Each encoding has a simple API inspired by Javascript's built-in `JSON` object:

```js
import { base32 } from "rfc4648";

base32.stringify([42, 121, 160]); // -> 'FJ42A==='
base32.parse("FJ42A==="); // -> Uint8Array([42, 121, 160])
```

The library has tree-shaking support, so tools like [rollup.js](https://rollupjs.org/) or [Webpack 2+](https://webpack.js.org/) can automatically trim away any encodings you don't use.

- Zero external dependencies
- 100% test coverage
- Built-in types for Typescript & Flow
- 0.8K minified + gzip (can be even smaller with tree shaking)

## API details

The library provides the following top-level modules:

- `base64`
- `base64url`
- `base32`
- `base32hex`
- `base16`
- `codec`

Each module exports a `parse` and `stringify` function.

### const string = baseXX.stringify(data, opts)

Each `stringify` function takes array-like object of bytes and returns a string.

If you pass the option `{ pad: false }` in the second parameter, the encoder will not output padding characters (`=`).

### const data = baseXX.parse(string, opts)

Each `parse` function takes a string and returns a `Uint8Array` of bytes. If you would like a different return type, such as plain `Array` or a Node.js `Buffer`, pass its constructor in the second argument:

```js
base64.parse("AOk=", { out: Array });
base64.parse("AOk=", { out: Buffer.allocUnsafe });
```

The constructor will be called with `new`, and should accept a single integer for the output length, in bytes.

If you pass the option `{ loose: true }` in the second parameter, the parser will not validate padding characters (`=`):

```js
base64.parse("AOk", { loose: true }); // No error
```

The base32 codec will also fix common typo characters in loose mode:

```js
base32.parse("He1l0==", { loose: true }); // Auto-corrects as 'HELLO==='
```

### Custom encodings

To define your own encodings, use the `codec` module:

```js
const codec = require("rfc4648").codec;

const myEncoding = {
  chars: "01234567",
  bits: 3
};

codec.stringify([220, 10], myEncoding); // '670050=='
codec.parse("670050", myEncoding, { loose: true }); // [ 220, 10 ]
```

The `encoding` structure should have two members, a `chars` member giving the alphabet and a `bits` member giving the bits per character. The `codec.parse` function will extend this with a third member, `codes`, the first time it's called. The `codes` member is a lookup table mapping from characters back to numbers.
