const {KEBConfig, KEBClient}= require('../kyma-environment-broker');
const {GardenerClient, GardenerConfig} = require("../gardener");
const {DirectorClient, DirectorConfig} = require("../compass");
const {genRandom, debug, getEnvOrThrow} = require("../utils");

const keb = new KEBClient(KEBConfig.fromEnv());
const gardener = new GardenerClient(GardenerConfig.fromEnv());
const director = new DirectorClient(DirectorConfig.fromEnv());

function WithRuntimeName(runtimeName) {
    return function (options) {
        options.runtimeName = runtimeName;
    }
}

function WithAppName(appName) {
    return function (options) {
        options.appName = appName;
    }
}

function WithScenarioName(scenarioName) {
    return function (options) {
        options.scenarioName = scenarioName;
    }
}

function WithTestNS(testNS) {
    return function (options) {
        options.testNS = testNS;
    }
}

function GatherOptions(...opts) {
    const suffix = genRandom(4);
    // If no opts provided the options object will be set to these default values.
    let options = {
        runtimeName: `kyma-${suffix}`,
        appName: `app-${suffix}`,
        scenarioName: `test-${suffix}`,
        testNS: "skr-test",
        // These options are not meant to be rewritten apart from env variable for KEB_USER_ID
        // If that's needed please add separate function that overrides this field.
        oidc0: {
            clientID: "abc-xyz",
            groupsClaim: "groups",
            issuerURL: "https://custom.ias.com",
            signingAlgs: ["RS256"],
            usernameClaim: "sub",
            usernamePrefix: "-",
        },
        oidc1: {
            clientID: "foo-bar",
            groupsClaim: "groups1",
            issuerURL: "https://new.custom.ias.com",
            signingAlgs: ["RS256"],
            usernameClaim: "email",
            usernamePrefix: "acme-",
        },
        administrator0: getEnvOrThrow("KEB_USER_ID"),
        administrators1: ["admin1@acme.com", "admin2@acme.com"],
    };

    opts.forEach((opt) => {
        opt(options);
    });
    debug(options);

    return options;
}

module.exports = {
    keb, gardener, director,
    GatherOptions,
    WithAppName,
    WithRuntimeName,
    WithScenarioName,
    WithTestNS,
}
