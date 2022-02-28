module.exports = {
    main: function (event, context) {
        console.log('Type of event.data: ', typeof event.data)
        console.log(event.data)
        console.log(event.datacontenttype)
    }
}