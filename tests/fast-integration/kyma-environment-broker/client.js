const axios = require("axios");
const {
    debug,
    getEnvOrThrow,
} = require("../utils");
const {
  OAuthCredentials,
  OAuthToken,
} = require("../lib/oauth");

const SCOPES = ["broker:write", "cld:read"];
const KYMA_SERVICE_ID = "47c9dcbf-ff30-448e-ab36-d3bad66ba281";

class KEBConfig {
  static fromEnv() {
    return new KEBConfig(
      getEnvOrThrow("KEB_HOST"),
      OAuthCredentials.fromEnv("KEB_CLIENT_ID", "KEB_CLIENT_SECRET"),
      getEnvOrThrow("KEB_GLOBALACCOUNT_ID"),
      getEnvOrThrow("KEB_SUBACCOUNT_ID"),
      getEnvOrThrow("KEB_USER_ID"),
      getEnvOrThrow("KEB_PLAN_ID"),
      process.env.KEB_REGION
    );
  }

  constructor(host, credentials, globalAccountID, subaccountID, userID, planID, region) {
    this.host = host;
    this.credentials = credentials;
    this.globalAccountID = globalAccountID;
    this.subaccountID = subaccountID;
    this.userID = userID;
    this.planID = planID;
    this.region = region;
  }
}


class KEBClient {
  constructor(config) {
    this.token = new OAuthToken(
      `https://oauth2.${config.host}/oauth2/token`, config.credentials);
    this.host = config.host;
    this.globalAccountID = config.globalAccountID;
    this.subaccountID = config.subaccountID;
    this.userID = config.userID;
    this.planID = config.planID
    this.region = config.region;
  }

  async buildRequest(payload, endpoint, verb) {
    const token = await this.token.getToken();
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

    try {
      const resp = await axios.request(config);
      if (resp.data.errors) {
        debug(resp);
        throw new Error(resp.data);
      }
      return resp.data;
    } catch (err) {
      debug(err);
      const msg = "Error calling KEB";
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
      service_id: KYMA_SERVICE_ID,
      plan_id: this.planID,
      context: {
        globalaccount_id: this.globalAccountID,
        subaccount_id: this.subaccountID,
        user_id: this.userID,
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
    const endpoint = `service_instances/${instanceID}?service_id=${KYMA_SERVICE_ID}&plan_id=${this.planID}`;
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
