const axios = require("axios");
const { debug } = require("../utils");

const kymaServiceID = "47c9dcbf-ff30-448e-ab36-d3bad66ba281";
class KEBClient {
  constructor(config) {
    this.host = config.kebHost;
    this.clientID = config.clientID;
    this.clientSecret = config.clientSecret;
    this.globalAccountID = config.globalAccountID;
    this.subAccountID = config.subAccountID;
    this.serviceID = kymaServiceID;
    this._token = void 0;
  }

  async getToken() {
    if (!this._token || this._token.expires_at < +new Date()) {
      const scopes = ["broker:write", "cld:read"];
      const url = `https://oauth2.${this.host}/oauth2/token`;
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
        this._token.expires_at = +new Date() + this._token.expires_in * 1000;
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
    const url = `https://kyma-env-broker.${this.host}/oauth/v2/${endpoint}`;
    debug(url);
    const headers = {
      "X-Broker-API-Version": 2.14,
      Authorization: `Bearer ${token}`,
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
        console.log(resp);
        console.log(resp.data.errors);
        throw new Error(resp.data);
      }
      return resp.data;
    } catch (err) {
      // console.dir(err);

      if (err.response) {
        debug(err.response);
        throw new Error(
          `${msg}: ${err.response.status} ${err.response.statusText}`
        );
      } else {
        throw new Error(`${msg}: ${err.toString()}`);
      }
    }
  }

  async provisionSKR(planID, name, instanceID) {
    const payload = {
      service_id: this.serviceID,
      plan_id: planID,
      context: {
        globalaccount_id: this.globalAccountID,
        subaccount_id: this.subAccountID,
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
    debug(endpoint)
    try {
      debug("CWEL")
      let dupa =  await this.callKEB({}, endpoint, "get");
      debug(dupa)
      return dupa
    } catch (err) {
      debug("SDADASDAS")
      debug(err.toString())
      return new Error(`error while checking SKR State: ${err.toString()}`);
    }
  }

  async deprovisionSKR(instanceID, planID) {
    const endpoint = `service_instances/${instanceID}?service_id=${this.serviceID}&plan_id=${planID}`;
    try {
      return await this.callKEB(null, endpoint, "delete");
    } catch (err) {
      return new Error(`error while deprovisioning SKR: ${err.toString()}`);
    }
  }

}

module.exports = {
  KEBClient,
};
