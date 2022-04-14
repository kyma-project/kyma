# durations

[![Build Status][travis-image]][travis-url]
[![NPM version][npm-image]][npm-url]

## Compatibilty

Both Node.js and browsers are supported by `durations`. When using Node.js, the nanosecond-granulatiry `process.hrtime()` function is used. The best substitution is selected when in the browser such that consistency is maintained even if time granularity cannot be.

## Installation

```shell
npm install --save durations
```

## Methods

The following functions are exported:
* `duration(nanoseconds)` - constructs a new Duration
* `nanos(nanoseconds)` - constructs a new Duration
* `micros(microseconds)` - constructs a new Duration
* `millis(milliseconds)` - constructs a new Duration
* `seconds(seconds)` - constructs a new Duration
* `stopwatch()` - constructs a new Stopwatch (stopped)
* `time(function)` - times a function synchronously
* `timeAsync(function(callback))` - times a function asynchronously
* `timePromised(function())` - times a promise-returning function

## Duration

Represents a duration with nanosecond granularity, and provides methods
for converting to other granularities, and formatting the duration.

### Methods
* `format()` - human readable string representing the duration
* `nanos()` - duration as nanoseconds
* `micros()` - duration as microseconds
* `millis()` - duration as milliseconds
* `seconds()` - duration as seconds
* `minutes()` - duration as minutes
* `hours()` - duration as hours
* `days()` - duration as days

```javascript
const {duration} = require('durations')

const nanoseconds = 987654321
console.log("Duration is", duration(nanoseconds).format())

// Or, since toString() is an alias to format()
console.log(`Duration is ${duration(nanoseconds)}`)
```

## Stopwatch

A nanosecond granularity (on Node.js) stopwatch with chainable control methods,
and built-in formatting.

### Stopwatch Methods
* `start()` - start and return the stopwatch (no-op if already running)
* `stop()` - stop and return the stopwatch (no-op if not running)
* `reset()` - reset to zero elapsed time and return the stopwatch (implies stop)
* `duration()` - fetch the elapsed time as a Duration
* `isRunning()` -  is the stopwatch running (`true`/`false`)

```javascript
const {stopwatch} = require('durations')
const watch = stopwatch()

// Pauses the stopwatch. Returns the stopwatch.
watch.stop()

// Starts the stopwatch from where it was last stopped. Returns the stopwatch.
watch.start()

// Reset the stopwatch (duration is set back to zero). Returns the stopwatch.
watch.reset()

console.log(`${watch.duration().seconds()} seconds have elapsed`)
// OR
console.log(`${watch} have elapsed`)
```

## Timer

Times the execution of a function, and returns the duration.

```javascript
const {time: timeSync, timeAsync} = require('durations')

// Synchronous work
const someFunction = () => {
  let count = 0

  while (count < 1000000) {
    count++
  }

  console.log(`Count is: ${count}`)
}

console.log(`Took ${timeSync(someFunction)} to do something`)

// Asynchronous work
const someOtherFunction = next => {
  someFunction()
  next()
}

timeAsync(someOtherFunction, duration => {
  console.log(`Took ${duration} to do something else.`)
})

// Promised work
const somePromisedOp = () => {
  return new Promise((resolve) => {
    someFunction()
    resolve()
  })
}

timePromised(somePromisedOp)
.then(duration => {
  console.log(`Took ${duration} to keep promise.`)
})
```

[travis-url]: https://travis-ci.org/joeledwards/node-durations
[travis-image]: https://img.shields.io/travis/joeledwards/node-durations/master.svg
[npm-url]: https://www.npmjs.com/package/durations
[npm-image]: https://img.shields.io/npm/v/durations.svg
