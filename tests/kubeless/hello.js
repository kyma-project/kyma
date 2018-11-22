'use strict';

const _ = require('lodash');

module.exports = {
    main: (event, context) => {
        console.log(event.data);
        var date = {
            'date': new Date().toTimeString()
        };
        if (_.isEmpty(event.data)) {
            _.assign(event.data, date);
        } else {
            var result = {
                'result': event.data
            };
            event.data = {};
            _.assignIn(event.data, result);
        }
        return JSON.stringify(event.data);
    },
};