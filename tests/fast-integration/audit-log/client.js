const axios = require("axios");
const moment = require('moment');
const {
    getEnvOrThrow,
} = require("../utils");

class AuditLogCreds {
    static fromEnv() {
        return JSON.parse(getEnvOrThrow("AL_SERVICE_KEY"))
    }
}

class AuditLogClient {
    constructor(credentials) { 
        this.creds = credentials
        this._token = undefined;
        this._logs = undefined
    }

    async getToken() {
        try {
            const resp = await axios.post(this.creds.uaa.url + "/oauth/token?grant_type=client_credentials", {},  {
                auth: {
                    username: this.creds.uaa.clientid,
                    password: this.creds.uaa.clientsecret
                }
            });
            this._token = resp.data.access_token;
            } 
        catch (err) {
            const msg = "Error when requesting bearer token from audit log";
            if (err.response) {
                throw new Error(
                `${msg}: ${err.response.status} ${err.response.statusText}`
                );
            } else {
                throw new Error(`${msg}: ${err.toString()}`);
            }
        }
        return this._token;
    }

    async fetchLogs() {
        const token = await this.getToken()
        const serviceUrl = this.creds.url
        let dateTo =  moment().utcOffset(0, false).format('YYYY-MM-DDTHH:mm:ss');
        let dateFrom = moment().utcOffset(0, false).subtract({'minutes': 10}).format('YYYY-MM-DDTHH:mm:ss');
        var url = serviceUrl + "/auditlog/v2/auditlogrecords?time_from=" + dateFrom + "&time_to=" + dateTo
        try {
            const resp = await axios.get(url, {
                    headers: {
                        "Authorization": `Bearer ${token}`
                    }
                })
            this._logs = resp.data
        }
        catch(err) {
            const msg = "Error when fetching logs from audit log service"
            if (err.response) {
                throw new Error(
                `${msg}: ${err.response.status} ${err.response.statusText}`
                );
            } else {
                throw new Error(`${msg}: ${err.toString()}`);
            }
        }
        return this._logs
    }
}

module.exports = {
    AuditLogCreds,
    AuditLogClient,
};
