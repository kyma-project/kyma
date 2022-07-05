const uuid = require('uuid');
const {genRandom, getEnvOrThrow, initializeK8sClient} = require('../utils');
const {keb, gardener, director} = require('./provision/provision-skr');
const {
  scenarioExistsInCompass,
  addScenarioInCompass,
  isRuntimeAssignedToScenario,
  assignRuntimeToScenario,
} = require('../compass');
const {saveKubeconfig} = require('../skr-svcat-migration-test/test-helpers');

const testNS = 'skr-test';

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
    customParams: {
      'name': 'dg-fit-test-1',
      'kymaVersion': 'PR-14747',
      'overridesVersion': '2.4.0-rc1',
    },
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

// gets the skr config by it's instance id
async function getSKRConfig(instanceID) {
  let shoot;
  try {
    shoot = await keb.getSKR(instanceID);
  } catch (e) {
    throw new Error(`Cannot fetch the shoot: ${e.toString()}`);
  }
  const shootName = shoot.dashboard_url.split('.')[1];

  console.log(`Fetching SKR info for shoot: ${shootName}`);
  return await gardener.getShoot(shootName);
}

async function prepareCompassResources(shoot, options) {
  // check if compass scenario setup is needed
  const compassScenarioAlreadyExist = await scenarioExistsInCompass(director, options.scenarioName);
  if (compassScenarioAlreadyExist) {
    console.log(`Compass scenario with the name ${options.scenarioName} already exist, do not register it again`);
  } else {
    console.log('Assigning SKR to scenario in Compass');
    // Create a new scenario (systems/formations) in compass for this test
    await addScenarioInCompass(director, options.scenarioName);
  }

  // check if assigning the runtime to the scenario is needed
  const runtimeAssignedToScenario = await isRuntimeAssignedToScenario(director,
      shoot.compassID,
      options.scenarioName);
  if (!runtimeAssignedToScenario) {
    console.log('Assigning Runtime to a compass scenario');
    // map scenario to target SKR
    await assignRuntimeToScenario(director, shoot.compassID, options.scenarioName);
  } else {
    console.log('Runtime %s is already assigned to the %s compass scenario', shoot.compassID, options.scenarioName);
  }
}

async function initK8sConfig(shoot) {
  console.log('Should save kubeconfig for the SKR to ~/.kube/config');
  await saveKubeconfig(shoot.kubeconfig);

  console.log('Should initialize K8s client');
  await initializeK8sClient({kubeconfig: shoot.kubeconfig});
}

module.exports = {
  testNS,
  getSKRConfig,
  prepareCompassResources,
  initK8sConfig,
  gatherOptions,
  withInstanceID,
  withAppName,
  withRuntimeName,
  withScenarioName,
  withTestNS,
  withSuffix,
  withCustomParams,
};
