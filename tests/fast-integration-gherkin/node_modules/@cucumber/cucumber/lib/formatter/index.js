"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
class Formatter {
    constructor(options) {
        this.colorFns = options.colorFns;
        this.cwd = options.cwd;
        this.eventDataCollector = options.eventDataCollector;
        this.log = options.log;
        this.snippetBuilder = options.snippetBuilder;
        this.stream = options.stream;
        this.supportCodeLibrary = options.supportCodeLibrary;
        this.cleanup = options.cleanup;
    }
    async finished() {
        await this.cleanup();
    }
}
exports.default = Formatter;
//# sourceMappingURL=index.js.map