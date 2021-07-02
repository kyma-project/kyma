const axios = require("axios");
const { 
    getEnvOrThrow 
} = require("../utils");

/**
 * Class OAuthCredentials wraps the OAuth credentials
 */
class OAuthCredentials {
    constructor(clientID, clientSecret) {
        this.clientID = clientID;
        this.clientSecret = clientSecret;
    }

    /**
     * Returns OAuthCredentials instance initialized from
     * the environment variables.
     * 
     * It expects the environment variables that store the credentials
     * to be present and not empty.
     * 
     * @param {string} clientIDEnv - client id environment variable name
     * @param {string} clientSecretEnv - client secret environment variable name
     * @returns {OAuthCredentials}
     */
    static fromEnv(clientIDEnv, clientSecretEnv) {
        return new OAuthCredentials(
            getEnvOrThrow(clientIDEnv),
            getEnvOrThrow(clientSecretEnv),
        );
    }
}

/**
 * Class OAuthToken provides simple approach to obtain and maintain
 * the OAuth2 token. 
 * 
 * This is very naive implementation just for the
 * internal fast-integration tests usage.
 */
class OAuthToken {
    constructor(url, credentials) {
        this.url = url;
        this.credentials = credentials;

        this._token = undefined;
    }

    async getToken(scopes) {
        if(!this._token || this._token.expires_at < +new Date()) {
            const body = `grant_type=client_credentials&scope=${scopes.join(" ")}`;
            const params = {
                auth: {
                    username: this.credentials.clientID,
                    password: this.credentials.clientSecret,
                },
                headers: {
                    "Content-Type": "application/x-www-form-urlencoded"
                }
            };

            try {
                const resp = await axios.post(this.url, body, params);
                this._token = resp.data;
                this._token.expores_at = (+new Date() + this._token.expires_in * 1000);
            } catch(err) {
                const msg = `Error when requesting bearer token from ${this.url}`;
                if(err.response) {
                    throw new Error(`${msg}: ${err.response.status} ${err.response.statusText}`);
                } else {
                    throw new Error(`${msg}: ${err.toString()}`);
                }
            }
        }

        return this._token.access_token;
    }
}

module.exports = {
    OAuthCredentials,
    OAuthToken,
};