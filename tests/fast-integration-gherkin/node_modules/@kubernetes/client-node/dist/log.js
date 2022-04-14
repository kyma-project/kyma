"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.Log = void 0;
const request = require("request");
const api_1 = require("./gen/api");
class Log {
    constructor(config) {
        this.config = config;
    }
    async log(namespace, podName, containerName, stream, doneOrOptions, options) {
        let done = () => undefined;
        if (typeof doneOrOptions === 'function') {
            done = doneOrOptions;
        }
        else {
            options = doneOrOptions;
        }
        const path = `/api/v1/namespaces/${namespace}/pods/${podName}/log`;
        const cluster = this.config.getCurrentCluster();
        if (!cluster) {
            throw new Error('No currently active cluster');
        }
        const url = cluster.server + path;
        const requestOptions = {
            method: 'GET',
            qs: {
                ...options,
                container: containerName,
            },
            uri: url,
        };
        await this.config.applyToRequest(requestOptions);
        return new Promise((resolve, reject) => {
            const req = request(requestOptions, (error, response, body) => {
                if (error) {
                    reject(error);
                    done(error);
                }
                else if (response.statusCode !== 200) {
                    try {
                        const deserializedBody = api_1.ObjectSerializer.deserialize(JSON.parse(body), 'V1Status');
                        reject(new api_1.HttpError(response, deserializedBody, response.statusCode));
                    }
                    catch (e) {
                        reject(new api_1.HttpError(response, body, response.statusCode));
                    }
                    done(body);
                }
                else {
                    done(null);
                }
            }).on('response', (response) => {
                if (response.statusCode === 200) {
                    req.pipe(stream);
                    resolve(req);
                }
            });
        });
    }
}
exports.Log = Log;
//# sourceMappingURL=log.js.map