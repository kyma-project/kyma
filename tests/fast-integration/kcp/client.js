const axios = require("axios");
const { OAuthCredentials, OAuthToken } = require("../lib/oauth");
const {getEnvOrThrow, debug} = require("../utils");

const SCOPES = ["cld:read"];

class KCPConfig {
    static fromEnv() {
        return new KCPConfig(
            getEnvOrThrow("KCP_HOST"),
            getEnvOrThrow("KCP_TECH_USER_LOGIN"),
            getEnvOrThrow("KCP_TECH_USER_PASSWORD"),
            OAuthCredentials.fromEnv("KCP_OIDC_CLIENT_ID", "KCP_OIDC_CLIENT_SECRET")
        );
    }
    constructor(host, login, password, credentials) {
        this.host = host;
        this.login = login;
        this.password = password;
        this.credentials = credentials;
    }
}

class KCPClient {
    constructor(config) {
        // TODO unhardcode it.
        // this url is required to get proper credentials for the KCP tech client.
        this.token = new OAuthToken(`https://kymatest.accounts400.ondemand.com`, config.credentials);
        this.host = config.host;
    }
    async buildRequest(payload, endpoint, verb) {
        const token = await this.token.getToken(SCOPES)
        const url = `https://kyma-env-broker.${this.host}/${endpoint}`
        const headers = {
            "accept": "application/json",
            Authorization: `Bearer ${token}`,
        }
        return {
            url: url,
            method: verb,
            headers: headers,
            data: payload,
        };
    }

    async runtimes(...customOptions) {
        let query;
        customOptions.forEach((option) => {
            if (option.account) {
                query = addParameter(query, `account`, option.account)
            }
            if (option.subaccount) {
                query = addParameter(query, `subaccount`, option.subaccount)
            }
            if (option.instanceID) {
                query = addParameter(query, `instance_id`, option.instanceID)
            }
            if (option.runtimeID) {
                query = addParameter(query, `runtime_id`, option.runtimeID)
            }
            if (option.region) {
                query = addParameter(query, `region`, option.region)
            }
            if (option.shoot) {
                query = addParameter(query, `shoot`, option.shoot)
            }
            if (option.state) {
                query = addParameter(query, `state`, option.state)
            }
        });

        const endpoint = `${this.host}/runtimes${query}`;
        console.log(endpoint)
        const req = await this.buildRequest({}, endpoint, "get")
        try {
            const resp = await axios.request(req);
            if (resp.data.error) {
                debug(resp)
                throw new Error(resp.data.error)
            }
            return resp.data
        } catch (err) {
            if (err.response) {
                throw new Error(`KEB get runtimes error: ${err.response.status} ${err.response.statusText}`);
            } else {
                throw new Error(`KEB get runtimes error: ${err.toString()}`);
            }
        }
    }
}

function addParameter(query, key, value) {
    if (query.length == 0) {
        query += '?';
    } else {
        query += '&';
    }
    query += `${key}=${value}`;
    return query
}

module.exports = {
    KCPConfig,
    KCPClient,
}
