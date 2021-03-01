const axios = require("axios");
const gql = require("./gql");

class ProvisionerClient {
    constructor(provisionerConfig) {
        this.host = provisionerConfig.provisionerHost;
        this.clientID = provisionerConfig.clientID;
        this.clientSecret = provisionerConfig.clientSecret;
        this.tenantID = provisionerConfig.tenantID;

        this._token = void 0;
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

    async callProvisioner(payload) {
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

        const msg = "Error calling Provisioner API"
        try {
            const resp = await axios.post(url, body, params);

            if(resp.data.errors) {
                console.log(resp)
                console.log(resp.data.errors);
                throw new Error(resp.data);
            }
            return resp.data.data.result;
        } catch(err) {
            // console.dir(err);
            if (err.response) {
                for(let e of err.response.data.errors) {
                    console.log(e);
                }
                throw new Error(`${msg}: ${err.response.status} ${err.response.statusText}`);
            } else if(err.errors) {
                throw new Error(`${msg}: GraphQL responded with errors: ${err.errors[0].message}`)
            } else {
                throw new Error(`${msg}: ${err.toString()}`);
            }
        }
    }

    async runtimeStatus(runtimeID) {
        const payload = gql.queryRuntimeStatus(runtimeID);
        try {
            return await this.callProvisioner(payload);
        } catch(err) {
            throw new Error(`Error when registering application: ${err.toString()}`);
        }
    }
}

module.exports = {
   ProvisionerClient
}; 