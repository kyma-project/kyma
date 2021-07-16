'use strict';

function newEventBase(event) {
    return Object.assign({}, event, {
        anyFunc1: (something) => helper1(something, event),
        anyFunc2: (something) => helper1(something, event)
    })
}

function helper1(something, somethingelse) {
    return "hello helper!"
}