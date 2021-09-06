const axios = require('axios');
const axiosRetry = require('axios-retry');
const https = require('https');

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
        const responseBody = await get(url);
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

async function get(url) {
    axiosRetry(axios, {
        retries: 5,
        retryDelay: (retryCount) => {
            return retryCount * 5000;
        },
        retryCondition: (error) => {
            console.log(error);
            return !error.response || error.response.status != 200;
        },
    });

    let response = await axios.get(url, {
        timeout: 5000,
    });
    return response;
}

module.exports = {
    lokiPortForward,
    queryLoki
};
