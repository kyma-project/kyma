#!/usr/bin/env coffee

isNode = if process?.hrtime()? then true else false

console.log "hrtime output is:", process.hrtime()
console.log "time is:", new Date().getTime()

console.log "Is Node.js: #{isNode}"

{stopwatch, time: timeSync, timeAsync} = require './src/index.coffee'

watch = stopwatch()
console.log "Duration should be zero:", watch.duration().nanos()
console.log "Formatted, no time registered: ", watch.duration().format()
console.log "Should be same format as above:", watch.format()

watch = stopwatch().start()
console.log "Duration should be non-zero:", watch.duration().nanos()
watch.stop()
console.log "Formatted duration, with time: ", watch.duration().format()
console.log "Should be same format as above:", watch.format()

console.log "Format on creation: ",
  stopwatch().start().stop().format()

action = ->
  num for num in [1 .. 5000000]

actionAsync = (next) ->
  num for num in [5000000 .. 10000000]
  next()

timeAsync(actionAsync, ((duration) ->
  console.log "Async timing:", duration.format()
))

console.log "Sync timing:", timeSync(action).format()

