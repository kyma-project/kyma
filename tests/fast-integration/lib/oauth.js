const { 
    getEnvOrThrow 
} = require("../utils");

class OAuthCredentials {
    constructor(clientID, clientSecret) {
        this.clientID = clientID;
        this.clientSecret = clientSecret;
    }

    static fromEnv(clientIDEnv, clientSecretEnv) {
        return new OAuthCredentials(
            getEnvOrThrow(clientIDEnv),
            getEnvOrThrow(clientSecretEnv),
        );
    }
}

module.exports = {
    OAuthCredentials,
};