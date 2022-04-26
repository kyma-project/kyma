const uuid = require('uuid');
const {KEBConfig, KEBClient}= require('../kyma-environment-broker');
const {GardenerClient, GardenerConfig} = require('../gardener');
const {DirectorClient, DirectorConfig} = require('../compass');
const {genRandom, debug, getEnvOrThrow} = require('../utils');
const execa = require("execa");

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
      clientID: '9bd05ed7-a930-44e6-8c79-e6defeb7dec5',
      groupsClaim: 'groups',
      issuerURL: 'https://kymatest.accounts400.ondemand.com',
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

async function smInstanceBinding(creds, btpOperatorInstance, btpOperatorBinding) {
  let args = [];
  try {
    args = ['login',
      '-a',
      creds.url,
      '--param',
      'subdomain=e2etestingscmigration',
      '--auth-flow',
      'client-credentials'];
    await execa('smctl', args.concat(['--client-id', creds.clientid, '--client-secret', creds.clientsecret]));

    args = ['provision', btpOperatorInstance, 'service-manager', 'service-operator-access', '--mode=sync'];
    await execa('smctl', args);

    // Move to Operator Install
    args = ['bind', btpOperatorInstance, btpOperatorBinding, '--mode=sync'];
    await execa('smctl', args);

    args = ['get-binding', btpOperatorBinding, '-o', 'json'];
    const out = await execa('smctl', args);
    const b = JSON.parse(out.stdout);
    const c = b.items[0].credentials;

    return {
      clientId: c.clientid,
      clientSecret: c.clientsecret,
      smURL: c.sm_url,
      url: c.url,
      instanceId: b.items[0].service_instance_id,
    };
  } catch (error) {
    if (error.stderr === undefined) {
      throw new Error(`failed to process output of "smctl ${args.join(' ')}"`);
    }
    throw new Error(`failed "smctl ${args.join(' ')}": ${error.stderr}`);
  }
}

class SMCreds {
  static fromEnv() {
    return new SMCreds(
        // TODO: rename to BTP_SM_ADMIN_CLIENTID
        getEnvOrThrow('BTP_OPERATOR_CLIENTID'),
        // TODO: rename to BTP_SM_ADMIN_CLIENTID
        getEnvOrThrow('BTP_OPERATOR_CLIENTSECRET'),
        // TODO: rename to BTP_SM_URL
        getEnvOrThrow('BTP_OPERATOR_URL'),
    );
  }

  constructor(clientid, clientsecret, url) {
    this.clientid = clientid;
    this.clientsecret = clientsecret;
    this.url = url;
  }
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
  smInstanceBinding,
  SMCreds,
};
