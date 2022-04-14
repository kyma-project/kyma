# ASN1.js

ASN.1 DER Encoder/Decoder and DSL for Node.js with no dependencies

## Acknowledgement

This is a fork of [@indutny's](https://github.com/indutny) `asn.js` library with the following
changes made:

- all `.int()` returns are native `BigInt` values
- lint using [`standard`](https://github.com/standard/standard)

## Example

Define model:

```js
const asn = require('@panva/asn1.js')

const Human = asn.define('Human', function () {
  this.seq().obj(
    this.key('firstName').octstr(),
    this.key('lastName').octstr(),
    this.key('age').int(),
    this.key('gender').enum({ 0: 'male', 1: 'female' }),
    this.key('bio').seqof(Bio)
  )
})

const Bio = asn.define('Bio', function () {
  this.seq().obj(
    this.key('time').gentime(),
    this.key('description').octstr()
  )
})
```

Encode data:

```js
const output = Human.encode({
  firstName: 'Thomas',
  lastName: 'Anderson',
  age: 28,
  gender: 'male',
  bio: [
    {
      time: new Date('31 March 1999').getTime(),
      description: 'freedom of mind'
    }
  ]
}, 'der')
```

Decode data:

```js
const human = Human.decode(output, 'der')
console.log(human)
/*
{ firstName: <Buffer 54 68 6f 6d 61 73>,
  lastName: <Buffer 41 6e 64 65 72 73 6f 6e>,
  age: 28n,
  gender: 'male',
  bio:
   [ { time: 922820400000,
       description: <Buffer 66 72 65 65 64 6f 6d 20 6f 66 20 6d 69 6e 64> } ] }
*/
```

### Partial decode

Its possible to parse data without stopping on first error. In order to do it,
you should call:

```js
const human = Human.decode(output, 'der', { partial: true })
console.log(human)
/*
{ result: { ... },
  errors: [ ... ] }
*/
```
