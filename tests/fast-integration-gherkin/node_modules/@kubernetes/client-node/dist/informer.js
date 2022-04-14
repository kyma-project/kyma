"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.makeInformer = exports.ERROR = exports.CONNECT = exports.DELETE = exports.CHANGE = exports.UPDATE = exports.ADD = void 0;
const cache_1 = require("./cache");
const watch_1 = require("./watch");
// These are issued per object
exports.ADD = 'add';
exports.UPDATE = 'update';
exports.CHANGE = 'change';
exports.DELETE = 'delete';
// This is issued when a watch connects or reconnects
exports.CONNECT = 'connect';
// This is issued when there is an error
exports.ERROR = 'error';
function makeInformer(kubeconfig, path, listPromiseFn, labelSelector) {
    const watch = new watch_1.Watch(kubeconfig);
    return new cache_1.ListWatch(path, watch, listPromiseFn, false, labelSelector);
}
exports.makeInformer = makeInformer;
//# sourceMappingURL=informer.js.map