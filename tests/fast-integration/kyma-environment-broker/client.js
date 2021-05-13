const axios = require("axios");
const {
    debug,
    getEnvOrThrow,
} = require("../utils");

const KYMA_SERVICE_ID = "47c9dcbf-ff30-448e-ab36-d3bad66ba281";

class KEBConfig {
  static fromEnv() {
    return new KEBConfig(
      getEnvOrThrow("KEB_HOST"),
      getEnvOrThrow("KEB_CLIENT_ID"),
      getEnvOrThrow("KEB_CLIENT_SECRET"),
      getEnvOrThrow("KEB_GLOBALACCOUNT_ID"),
      getEnvOrThrow("KEB_SUBACCOUNT_ID"),
      getEnvOrThrow("KEB_PLAN_ID"),
      process.env.KEB_REGION
    );
  }

  constructor(host, clientID, clientSecret, globalAccountID, subaccountID, planID, region) {
    this.host = host;
    this.clientID = clientID;
    this.clientSecret = clientSecret;
    this.globalAccountID = globalAccountID;
    this.subaccountID = subaccountID;
    this.planID = planID;
    this.region = region;
  }
}


class KEBClient {
  constructor(config) {
    this.host = config.host;
    this.clientID = config.clientID;
    this.clientSecret = config.clientSecret;
    this.globalAccountID = config.globalAccountID;
    this.subaccountID = config.subaccountID;
    this.planID = config.planID
    this.serviceID = KYMA_SERVICE_ID;
    this.region = config.region;

    this._token = undefined;
  }

  async getToken() {
    if (!this._token || this._token.expires_at < +new Date()) {
      const url = `https://oauth2.${this.host}/oauth2/token`;
      const scopes = ["broker:write", "cld:read"];
      const body = `grant_type=client_credentials&scope=${scopes.join(" ")}`;
      const params = {
        auth: {
          username: this.clientID,
          password: this.clientSecret,
        },
      };

      try {
        const resp = await axios.post(url, body, params);

        this._token = resp.data;
        this._token.expires_at = (+new Date() + this._token.expires_in * 1000);
      } catch (err) {
        const msg = "Error when requesting bearer token from KCP";
        if (err.response) {
          throw new Error(
            `${msg}: ${err.response.status} ${err.response.statusText}`
          );
        } else {
          throw new Error(`${msg}: ${err.toString()}`);
        }
      }
    }

    return this._token.access_token;
  }

  async buildRequest(payload, endpoint, verb) {
    const token = await this.getToken();
    const region = this.getRegion();
    const url = `https://kyma-env-broker.${this.host}/oauth/${region}v2/${endpoint}`;
    const headers = {
      "X-Broker-API-Version": 2.14,
      "Authorization": `Bearer ${token}`,
      "Content-Type": "application/json",
    };

    const request = {
      url: url,
      method: verb,
      headers: headers,
      data: payload,
    };
    return request;
  }

  async callKEB(payload, endpoint, verb) {
    const config = await this.buildRequest(payload, endpoint, verb);

    const msg = "Error calling KEB";
    try {
      const resp = await axios.request(config);
      if (resp.data.errors) {
        debug(resp);
        throw new Error(resp.data);
      }
      return resp.data;
    } catch (err) {
      debug(err);
      if (err.response) {
        throw new Error(
          `${msg}: ${err.response.status} ${err.response.statusText}`
        );
      } else {
        throw new Error(`${msg}: ${err.toString()}`);
      }
    }
  }

  async provisionSKR(name, instanceID) {
    const payload = {
      service_id: this.serviceID,
      plan_id: this.planID,
      context: {
        globalaccount_id: this.globalAccountID,
        subaccount_id: this.subaccountID,
      },
      parameters: {
        name: name,
      },
    };
    const endpoint = `service_instances/${instanceID}`;
    try {
      return await this.callKEB(payload, endpoint, "put");
    } catch (err) {
      throw new Error(`error while provisioning SKR: ${err.toString()}`);
    }
  }

  async getOperation(instanceID, operationID) {
    const endpoint = `service_instances/${instanceID}/last_operation?operation=${operationID}`;
    try {
      return await this.callKEB({}, endpoint, "get");
    } catch (err) {
      debug(err.toString())
      return new Error(`error while checking SKR State: ${err.toString()}`);
    }
  }

  async deprovisionSKR(instanceID) {
    const endpoint = `service_instances/${instanceID}?service_id=${this.serviceID}&plan_id=${this.planID}`;
    try {
      return await this.callKEB(null, endpoint, "delete");
    } catch (err) {
      return new Error(`error while deprovisioning SKR: ${err.toString()}`);
    }
  }

  getRegion() {
    if (this.region && this.region != "") {
      return `${this.region}/`;
    }
    return "";
  }
}

module.exports = {
  KEBConfig,
  KEBClient,
};
