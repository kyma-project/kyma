const axios = require('axios');

const {
    kubectlPortForward,
} = require("../utils");

const lokiPort = 3100;

function lokiPortForward() {
    return kubectlPortForward("kyma-system", "logging-loki-0", lokiPort);
}

async function queryLoki(labels, startTimestamp) {
    try {
        const url = `http://localhost:${lokiPort}/api/prom/query?query=${labels}&start=${startTimestamp}`;
        const responseBody = await axios.get(url);
        return responseBody.data;
    } catch(err) {
        const msg = "Error when querying Loki";
        if (err.response) {
            throw new Error(`${msg}: ${err.response.status} ${err.response.statusText}: ${err.response.data}`);
        } else {
            throw new Error(`${msg}: ${err.toString()}`);
        }
    }
}

module.exports = {
    lokiPortForward,
    queryLoki
};
