const uuid = require('uuid');
const {KEBConfig, KEBClient}= require('../kyma-environment-broker');
const {GardenerClient, GardenerConfig} = require('../../gardener');
const {KCPWrapper, KCPConfig} = require('../kcp/client');
const fs = require('fs');
const os = require('os');
const {expect} = require('chai');

const keb = new KEBClient(KEBConfig.fromEnv());
const gardener = new GardenerClient(GardenerConfig.fromEnv());
const kcp = new KCPWrapper(KCPConfig.fromEnv());
const testNS = 'skr-test';
const DEBUG = process.env.DEBUG === 'true';
const {initializeK8sClient} = require('../../utils/index.js');

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

function withSuffix(suffix) {
  return function(options) {
    options.suffix = suffix;
  };
}

function withCustomParams(customParams) {
  return function(options) {
    options.customParams = customParams;
  };
}

function gatherOptions(...opts) {
  // If no opts provided the options object will be set to these default values.
  const options = {
    instanceID: uuid.v4(),
    testNS: testNS,
    // These options are not meant to be rewritten apart from env variable for KEB_USER_ID
    // If that's needed please add separate function that overrides this field.
    oidc0: {
      clientID: '9bd05ed7-a930-44e6-8c79-e6defeb7dec9',
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
    kebUserId: getEnvOrThrow('KEB_USER_ID'),
    administrators1: ['admin1@acme.com', 'admin2@acme.com'],
    customParams: {},
  };

  opts.forEach((opt) => {
    opt(options);
  });

  if (options.suffix === undefined) {
    options.suffix = genRandom(4);
  }

  options.runtimeName = `kyma-${options.suffix}`;
  options.appName = `app-${options.suffix}`;
  options.scenarioName = `test-${options.suffix}`;

  return options;
}

async function getSKRRuntimeStatus(instanceID) {
  const runtimeStatus = await kcp.getRuntimeStatusOperations(instanceID);
  return JSON.parse(runtimeStatus);
}

// gets the skr config by its instance id
async function getSKRConfig(instanceID) {
  const runtimeStatus = await kcp.getRuntimeStatusOperations(instanceID);
  const objRuntimeStatus = JSON.parse(runtimeStatus);
  expect(objRuntimeStatus).to.have.nested.property('data[0].shootName').not.empty;
  const shootName = objRuntimeStatus.data[0].shootName;

  console.log(`Fetching SKR info for shoot: ${shootName}`);
  return await gardener.getShoot(shootName);
}

function getEnvOrThrow(key) {
  if (!process.env[key]) {
    throw new Error(`Env ${key} not present`);
  }

  return process.env[key];
}

function genRandom(len) {
  let res = '';
  const chrs = 'abcdefghijklmnopqrstuvwxyz0123456789';
  for (let i = 0; i < len; i++) {
    res += chrs.charAt(Math.floor(Math.random() * chrs.length));
  }

  return res;
}

function debug(...args) {
  if (!DEBUG) {
    return;
  }
  log('DEBUG', ...args);
}

function log(prefix, ...args) {
  if (args.length === 0) {
    return;
  }

  args = [...args];
  const fmt = `[${prefix}] ` + args[0];
  args = args.slice(1);
  console.log.apply(console, [fmt, ...args]);
}

async function saveKubeconfig(kubeconfig) {
  if (!fs.existsSync(`${os.homedir()}/.kube`)) {
    fs.mkdirSync(`${os.homedir()}/.kube`, true);
  }
  fs.writeFileSync(`${os.homedir()}/.kube/config`, kubeconfig);
}

async function initK8sConfig(shoot) {
  console.log('Should save kubeconfig for the SKR to ~/.kube/config');
  await saveKubeconfig(shoot.kubeconfig);

  console.log('Should initialize K8s client');
  await initializeK8sClient({kubeconfig: shoot.kubeconfig});
}

module.exports = {
  keb,
  kcp,
  gardener,
  getSKRConfig,
  getSKRRuntimeStatus,
  gatherOptions,
  withInstanceID,
  withAppName,
  withRuntimeName,
  withScenarioName,
  withTestNS,
  withSuffix,
  withCustomParams,
  getEnvOrThrow,
  genRandom,
  debug,
  saveKubeconfig,
  log,
  initK8sConfig,
};
