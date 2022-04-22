const uuid = require('uuid');
const {KEBConfig, KEBClient}= require('../kyma-environment-broker');
const {GardenerClient, GardenerConfig} = require('../gardener');
const {DirectorClient, DirectorConfig} = require('../compass');
const {genRandom, debug, getEnvOrThrow} = require('../utils');

const keb = new KEBClient(KEBConfig.fromEnv());
const gardener = new GardenerClient(GardenerConfig.fromEnv());
const director = new DirectorClient(DirectorConfig.fromEnv());

function withInstanceID(instanceID) {
  return function(options) {
    options.instanceID = instanceID;
  };
}

function withRuntimeName(runtimeName) {
  return function(options) {
    options.runtimeName = runtimeName;
  };
}

function withAppName(appName) {
  return function(options) {
    options.appName = appName;
  };
}

function withScenarioName(scenarioName) {
  return function(options) {
    options.scenarioName = scenarioName;
  };
}

function withTestNS(testNS) {
  return function(options) {
    options.testNS = testNS;
  };
}

function gatherOptions(...opts) {
  const suffix = genRandom(4);
  // If no opts provided the options object will be set to these default values.
  const options = {
    instanceID: uuid.v4(),
    runtimeName: `kyma-${suffix}`,
    appName: `app-${suffix}`,
    scenarioName: `test-${suffix}`,
    testNS: 'skr-test',
    // These options are not meant to be rewritten apart from env variable for KEB_USER_ID
    // If that's needed please add separate function that overrides this field.
    oidc0: {
      clientID: 'abc-xyz',
      groupsClaim: 'groups',
      issuerURL: 'https://custom.ias.com',
      signingAlgs: ['RS256'],
      usernameClaim: 'sub',
      usernamePrefix: '-',
    },
    oidc1: {
      clientID: 'foo-bar',
      groupsClaim: 'groups1',
      issuerURL: 'https://new.custom.ias.com',
      signingAlgs: ['RS256'],
      usernameClaim: 'email',
      usernamePrefix: 'acme-',
    },
    administrator0: getEnvOrThrow('KEB_USER_ID'),
    administrators1: ['admin1@acme.com', 'admin2@acme.com'],
  };

  opts.forEach((opt) => {
    opt(options);
  });
  debug(options);

  return options;
}

module.exports = {
  keb,
  gardener,
  director,
  gatherOptions,
  withInstanceID,
  withAppName,
  withRuntimeName,
  withScenarioName,
  withTestNS,
};
