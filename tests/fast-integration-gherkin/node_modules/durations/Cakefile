{spawn, exec} = require 'child_process'

launch = (cmd, options=[], done=null) ->
    app = spawn cmd, options
    app.stdout.pipe(process.stdout)
    app.stderr.pipe(process.stderr)
    app.on 'exit', (status) ->
        err = if status isnt 0 then new Error("Error running #{cmd}") else null
        done? err

build = (done) ->
    console.log "Building"
    exec './node_modules/.bin/coffee --compile --output lib/ src/', (err, stdout, stderr) ->
        process.stderr.write stderr
        return done err if err

        process.stderr.write stderr
        done?()

mocha = (done) ->
    console.log "Testing"
    launch './node_modules/.bin/mocha', ['test'], done

run = (fn) ->
    ->
        fn (err) ->
            console.log err.stack if err

task 'build', "Build project from src/*.coffee to lib/*.js", run build
task 'test', "Run mocha tests", run -> build mocha
