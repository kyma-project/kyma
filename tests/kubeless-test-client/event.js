module.exports = {
    handler: function (event, context) {
        console.log("OK", event.data);
        return event.data;
    }
}
