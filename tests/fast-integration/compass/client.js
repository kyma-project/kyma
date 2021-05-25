const axios = require("axios");
const gql = require("./gql");
const { getEnvOrThrow, debug } = require("../utils")

/**
 * Class DirectorConfig represents configuration data for DirectorClient.
 */
class DirectorConfig {
    /**
     * Returns DirectorConfig class initialized from
     * the environment variables.
     * 
     * Expects the following variables to be present:
     * - COMPASS_HOST
     * - COMPASS_CLIENT_ID
     * - COMPASS_CLIENT_SECRET
     * - COMPASS_TENANT
     * 
     * @returns {DirectorConfig}
     */
    static fromEnv() {
        return new DirectorConfig(
            getEnvOrThrow("COMPASS_HOST"),
            getEnvOrThrow("COMPASS_CLIENT_ID"),
            getEnvOrThrow("COMPASS_CLIENT_SECRET"),
            getEnvOrThrow("COMPASS_TENANT")
        )
    }

    constructor(host, clientID, clientSecret, tenantID) {
        this.host = host;
        this.clientID = clientID;
        this.clientSecret = clientSecret;
        this.tenantID = tenantID;
    }
}

/**
 * Class DirectorClient represents API methods of the Director component.
 */
class DirectorClient {
    /**
     * Create a DirectorClient instance.
     * 
     * @param {DirectorConfig} config 
     */
    constructor(config) {
        this.host = config.host;
        this.clientID = config.clientID;
        this.clientSecret = config.clientSecret;
        this.tenantID = config.tenantID;

        this._token = undefined;
    }

    async getToken() {
        if (!this._token || this._token.expires_at < +new Date()) {
            const scopes = [
                "application:read",
                "application:write", 
                "runtime:read",
                "runtime:write", 
                "label_definition:read", 
                "label_definition:write"
            ];
            const url = `https://oauth2.${this.host}/oauth2/token`;
            const body = `grant_type=client_credentials&scope=${scopes.join(" ")}`;
            const params = {
                auth: {
                    username: this.clientID,
                    password: this.clientSecret
                },
                headers: {
                    "Content-Type": "application/x-www-form-urlencoded"
                }
            };
        
            try {
                const resp = await axios.post(url, body, params);
                
                this._token = resp.data;
                this._token.expires_at = (+new Date() + this._token.expires_in * 1000);
            } catch(err) {
                const msg = "Error when requesting bearer token from compass"
                if (err.response) {
                    throw new Error(`${msg}: ${err.response.status} ${err.response.statusText}`);
                } else {
                    throw new Error(`${msg}: ${err.toString()}`);
                }
            }
        }
        
        return this._token.access_token;
    }

    async callDirector(payload) {
        const token = await this.getToken();
        const url = `https://compass-gateway-auth-oauth.${this.host}/director/graphql`
        const body = `{"query":"${payload}"}`;
        const params = {
            headers: {
                "Tenant": this.tenantID,
                "Authorization": `Bearer ${token}`,
                "Content-Type": "application/json"
            }
        };

        const msg = "Error calling Director API"
        try {
            const resp = await axios.post(url, body, params);
            if(resp.data.errors) {
                debug(resp);
                console.log(`DEBUG callDirector resp: ${resp}`);
                console.log(`DEBUG callDirector resp: ${JSON.stringify(resp)}`);
                console.error(resp);
                console.error(JSON.stringify(resp));
                throw new Error(resp.data);
            }
            return resp.data.data.result;
        } catch(err) {
            debug(err);
            console.log(`DEBUG callDirector err: ${err}`);
            console.log(`DEBUG callDirector err: ${JSON.stringify(err)}`);
            console.error(err);
            console.error(JSON.stringify(err));
            if (err.response) {
                throw new Error(`${msg}: ${err.response.status} ${err.response.statusText}`);
            } else if(err.errors) {
                throw new Error(`${msg}: GraphQL responded with errors: ${err.errors[0].message}`)
            } else {
                throw new Error(`${msg}: ${err.message}`);
            }
        }
    }

    async registerApplication(appName, scenarioName) {
        const payload = gql.registerApplication(appName, scenarioName);
        try {
            const res = await this.callDirector(payload);
            return res.id;
        } catch(err) {
            throw new Error(`Error when registering application: ${err.toString()}`);
        }
    }

    async unregisterApplication(applicationID) {
        const payload = gql.unregisterApplication(applicationID);
        try {
            await this.callDirector(payload);
        } catch(err) {
            throw new Error(`Error when unregistering application: ${err.message}`);
        }
    }

    async registerRuntime(runtimeName, scenarioName) {
        const payload = gql.registerRuntime(runtimeName, scenarioName);
        try {
            const res = await this.callDirector(payload);
            return res.id;
        } catch(err) {
            throw new Error(`Error when registering runtime: ${err.toString()}`);
        }
    }

    async unregisterRuntime(runtimeID) {
        const payload = gql.unregisterRuntime(runtimeID);
        try {
            await this.callDirector(payload);
        } catch(err) {
            throw new Error(`Error when unregistering runtime: ${err.toString()}`);
        }
    }

    async requestOneTimeTokenForApplication(appID) {
        const payload = gql.requestOneTimeTokenForApplication(appID);
        try {
            const res = await this.callDirector(payload);
            return res; // {token: '...', connectorURL: '...'}
        } catch(err) {
            throw new Error(`Error when requesting token for application: ${err.toString()}`);
        }
    }

    async requestOneTimeTokenForRuntime(runtimeID) {
        const payload = gql.requestOneTimeTokenForRuntime(runtimeID);
        try {
            const res = await this.callDirector(payload);
            return res; // {token: '...', connectorURL: '...'}
        } catch(err) {
            throw new Error(`Error when requesting token for runtime: ${err.toString()}`);
        }
    }

    async queryLabelDefinition(labelKey) {
        const payload = gql.queryLabelDefinition(labelKey);
        try {
            const res = await this.callDirector(payload);
            if (res.schema) {
                res.schema = JSON.parse(res.schema);
            }
            return res;
        } catch(err) {
            throw new Error(`Error when querying for label definition with key ${labelKey}: ${err.toString()}`);
        }
    }

    async updateLabelDefinition(labelKey, schema) {
        const payload = gql.updateLabelDefinition(labelKey, schema);
        try {
            await this.callDirector(payload);
        } catch(err) {
            throw new Error(`Error when updating label definition with key ${labelKey}: ${err.toString()}`);
        }
    }

    async queryRuntimesWithFilter(filter) {
        const payload = gql.queryRuntimesWithFilter(filter);
        try {
            const res = await this.callDirector(payload);
            return res.data;
        } catch(err) {
            throw new Error(`Error when querying for runtimes filtered: ${err.toString()}`);
        }
    }

    async queryApplicationsWithFilter(filter) {
        const payload = gql.queryApplicationsWithFilter(filter);
        try {
            const res = await this.callDirector(payload);
            return res.data;
        } catch(err) {
            throw new Error(`Error when querying for applications filtered: ${err.toString()}`);
        }
    }

    async setRuntimeLabel(runtimeID, key, value) {
        const payload = gql.setRuntimeLabel(runtimeID, key, value);
        try {
            const res = await this.callDirector(payload);
            return res.data;
        } catch(err) {
            throw new Error(`Error when setting runtime ${runtimeID} label ${key} and value ${value}: ${err.toString()}`);
        }
    }

    async getRuntime(runtimeID) {
        const payload = gql.queryRuntime(runtimeID);
        try {
            const res = await this.callDirector(payload);
            return res;
        } catch(err) {
            throw new Error(`Error whe querying for the runtime with ID ${runtimeID}: ${err.toString()}`);
        }
    }

    async getApplication(appID) {
        const payload = gql.queryApplication(appID);
        try {
            const res = await this.callDirector(payload);
            return res;
        } catch(err) {
            throw new Error(`Error when querying for the application with ID ${appID}: ${err.toString()}`);
        }
    }

    async setApplicationLabel(appID, key, value) {
        const payload = gql.setApplicationLabel(appID, key, value);
        try {
            const res = await this.callDirector(payload);
            return res.data;
        } catch(err) {
            throw new Error(`Error when setting application ${appID} label ${key} and value ${value}: ${err.toString()}`);
        }
    }

    async deleteApplicationLabel(appID, key) {
        const payload = gql.deleteApplicationLabel(appID, key);
        try {
            const res = await this.callDirector(payload);
            return res.data;
        } catch(err) {
            throw new Error(`Error when deleting label ${key} from application ${appID}: ${err.toString()}`);
        }
    }
}

module.exports = {
    DirectorConfig,
    DirectorClient
};