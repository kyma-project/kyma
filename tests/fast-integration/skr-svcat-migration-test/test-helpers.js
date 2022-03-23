const execa = require('execa');
const fs = require('fs');
const os = require('os');
const {
  getEnvOrThrow,
  getConfigMap,
  kubectlExecInPod,
  listPods,
  deleteK8sPod,
  sleep,
  deleteK8sObjects,
  listResources,
  getFunction,
} = require('../utils');

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

const functions = [
  {name: 'svcat-auditlog-api-1', checkEnvVars: 'uaa url vendor'},
  {name: 'svcat-auditlog-management-1', checkEnvVars: 'uaa url vendor'},
  {name: 'svcat-html5-apps-repo-1', checkEnvVars: 'grant_type saasregistryenabled sap.cloud.service uaa uri vendor'},
];

async function saveKubeconfig(kubeconfig) {
  if (!fs.existsSync(`${os.homedir()}/.kube`)) {
    fs.mkdirSync(`${os.homedir()}/.kube`, true);
  }
  fs.writeFileSync(`${os.homedir()}/.kube/config`, kubeconfig);
}

async function readClusterID() {
  const cm = await getConfigMap('cluster-info', 'kyma-system');
  return cm.data.id;
}

async function functionReady(functionName) {
  const fn = await getFunction(functionName, 'default');
  return fn.status.conditions.reduce((acc, val) => acc && val.status == 'True', true);
}

async function getFunctionPod(functionName) {
  const labelSelector = `serverless.kyma-project.io/function-name=${functionName},` +
    'serverless.kyma-project.io/resource=deployment';
  let res = {};
  for (let i = 0; i < 30; i++) {
    const ready = await functionReady(functionName);
    if (ready) {
      res = await listPods(labelSelector);
      if (res.body.items.length == 1) {
        const pod = res.body.items[0];
        if (pod.status.phase == 'Running') {
          return pod;
        }
      }
    }
    sleep(10000);
  }
  const podNames = res.body.items.map((p) => p.metadata.name);
  const phases = res.body.items.map((p) => p.status.phase);
  const fn = await getFunction(functionName, 'default');
  throw new Error(`Failed to find function ${functionName} pod in 5 minutes.
  Expected 1 ${labelSelector} pod with phase "Running" but found ${res.body.items.length}, ${podNames}, ${phases}\n
  function status: ${JSON.stringify(fn.status)}`);
}

async function checkPodPresetEnvInjected() {
  const cmd = 'for v in {vars}; do x="$(eval echo \\$$v)"; if [[ -z "$x" ]];' +
    'then echo missing $v env variable; exit 1; else echo found $v env variable; fi; done';
  for (const f of functions) {
    const pod = await getFunctionPod(f.name);
    const envCmd = cmd.replace('{vars}', f.checkEnvVars);
    await kubectlExecInPod(pod.metadata.name, 'function', ['sh', '-c', envCmd]);
  }
}

async function restartFunctionsPods() {
  const podNames = {};
  for (const f of functions) {
    const pod = await getFunctionPod(f.name);
    console.log('delete pod', pod.metadata.name);
    try {
      await deleteK8sPod(pod);
    } catch (err) {
      throw new Error(`failed to delete pod ${pod.metadata.name}: ${err}`);
    }
    podNames[f.name] = pod.metadata.name;
  }

  let needsPoll = [];
  for (let i = 0; i < 10; i++) {
    needsPoll = [];
    for (const f of functions) {
      const labelSelector = `serverless.kyma-project.io/function-name=${f.name},` +
        'serverless.kyma-project.io/resource=deployment';
      console.log(`polling pods with labelSelector ${labelSelector}`);
      let res = {};
      try {
        res = await listPods(labelSelector);
      } catch (err) {
        throw new Error(`failed to list pods with labelSelector ${labelSelector}: ${err}`);
      }
      if (res.body.items.length != 1) {
        // there are either multiple or 0 pods for the function, we need to wait
        const pn = res.body.items.map((p) => {
          return {'pod name': p.metadata.name, 'phase': p.status.phase};
        });
        needsPoll.push({'function name': f.name, 'pods': pn});
        continue;
      }
      const pod = res.body.items[0];
      const pn = pod.metadata.name;
      const psp = pod.status.phase;
      if (pn == podNames[f.name]) {
        // there is single pod for the function, but it is still the old one
        needsPoll.push({'function name': f.name, 'pods': [{'pod name': pn, 'phase': psp}]});
        continue;
      }
      if (psp != 'Running') {
        // there is single pod for the function, it has new name, but it's not in Running state
        needsPoll.push({'function name': f.name, 'pods': [{'pod name': pn, 'phase': psp}]});
        continue;
      }
    }
    if (needsPoll.length != 0) {
      await sleep(10000); // 10 seconds
    } else {
      break;
    }
  }
  if (needsPoll.length != 0) {
    const info = JSON.stringify(needsPoll, null, 2);
    const originalNames = JSON.stringify(podNames, null, 2);
    throw new Error(`Failed to restart function pods in 100 seconds.
    Expecting exactly one pod for each function with new unique names and in ready status but found:
    ${info}
    Pod names before restart:
    ${originalNames}`);
  }
  console.log('functions pods successfully restarted');
}

async function provisionPlatform(creds, svcatPlatform) {
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

    // $ smctl register-platform <name> kubernetes -o json
    // Output:
    // {
    //   "id": "<platform-id/cluster-id>",
    //   "name": "<name>",
    //   "type": "kubernetes",
    //   "created_at": "...",
    //   "updated_at": "...",
    //   "credentials": {
    //     "basic": {
    //       "username": "...",
    //       "password": "..."
    //     }
    //   },
    //   "labels": {
    //     "subaccount_id": [
    //       "..."
    //     ]
    //   },
    //   "ready": true
    // }
    args = ['register-platform', svcatPlatform, 'kubernetes', '-o', 'json'];
    const registerPlatformOut = await execa('smctl', args);
    const platform = JSON.parse(registerPlatformOut.stdout);

    return {
      clusterId: platform.id,
      name: platform.name,
      credentials: platform.credentials.basic,
    };
  } catch (error) {
    if (error.stderr === undefined) {
      throw new Error(`failed to process output of "smctl ${args.join(' ')}"`);
    }
    throw new Error(`failed "smctl ${args.join(' ')}": ${error.stderr}`);
  }
}

async function smInstanceBinding(btpOperatorInstance, btpOperatorBinding) {
  let args = [];
  try {
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

async function markForMigration(creds, svcatPlatform, btpOperatorInstanceId) {
  let errors = [];
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
  } catch (error) {
    errors = errors.concat([`failed "smctl ${args.join(' ')}": ${error.stderr}\n`]);
  }

  try {
    // usage: smctl curl -X PUT -d '{"sourcePlatformID": ":platformID"}' /v1/migrate/service_operator/:instanceID
    const data = {sourcePlatformID: svcatPlatform};
    args = ['curl', '-X', 'PUT', '-d', JSON.stringify(data), '/v1/migrate/service_operator/' + btpOperatorInstanceId];
    await execa('smctl', args);
  } catch (error) {
    errors = errors.concat([`failed "smctl ${args.join(' ')}": ${error.stderr}\n`]);
  }
  if (errors.length > 0) {
    throw new Error(errors.join(', '));
  }
}

async function cleanupInstanceBinding(creds, svcatPlatform, btpOperatorInstance, btpOperatorBinding) {
  let errors = [];
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
  } catch (error) {
    errors = errors.concat([`failed "smctl ${args.join(' ')}": ${error.stderr}\n`]);
  }

  try {
    args = ['unbind', btpOperatorInstance, btpOperatorBinding, '-f', '--mode=sync'];
    const {stdout} = await execa('smctl', args);
    if (stdout !== 'Service Binding successfully deleted.') {
      errors = errors.concat([`failed "smctl ${args.join(' ')}": ${stdout}`]);
    }
  } catch (error) {
    errors = errors.concat([`failed "smctl ${args.join(' ')}": ${error.stderr}\n${error}`]);
  }

  try {
    // hint: probably should fail cause that instance created other instannces (after the migration is done)
    args = ['deprovision', btpOperatorInstance, '-f', '--mode=sync'];
    const {stdout} = await execa('smctl', args);
    if (stdout !== 'Service Instance successfully deleted.') {
      errors = errors.concat([`failed "smctl ${args.join(' ')}": ${stdout}`]);
    }
  } catch (error) {
    errors = errors.concat([`failed "smctl ${args.join(' ')}": ${error.stderr}\n${error}`]);
  }

  try {
    args = ['delete-platform', svcatPlatform, '-f', '--cascade'];
    await execa('smctl', args);
    // if (stdout !== "Platform(s) successfully deleted.") {
    //     errors = errors.concat([`failed "smctl ${args.join(' ')}": ${stdout}`])
    // }
  } catch (error) {
    errors = errors.concat([`failed "smctl ${args.join(' ')}": ${error.stderr}\n`]);
  }

  if (errors.length > 0) {
    throw new Error(errors.join(', '));
  }
}

async function checkMigratedBTPResources() {
  const btpGroup = 'services.cloud.sap.com';
  const btpVersion = 'v1alpha1';
  const scGroup = 'servicecatalog.k8s.io';
  const scVersion = 'v1beta1';
  const instances = 'serviceinstances';
  const bindings = 'servicebindings';

  let errors = [];
  // the test shouldn't wait here too long, it's just elementary re-try to reduce flakiness
  // the reconciler is supposed to keep the update operation in pending for the time of migrating and cleanup
  for (let i = 0; i < 3; i++) {
    errors = [];
    const btpBindings = await listResources(`/apis/${btpGroup}/${btpVersion}/${bindings}`);
    const bindingsReady = btpBindings.reduce((ready, binding) => ready && binding.status.ready === 'True', true);
    if (btpBindings.length != 3 || !bindingsReady) {
      const bs = JSON.stringify(btpBindings, null, 2);
      errors.push(`Expected 3 BTP bindings ready but found ${btpBindings.length}:\n${bs}`);
    }
    const btpInstances = await listResources(`/apis/${btpGroup}/${btpVersion}/${instances}`);
    const instancesReady = btpInstances.reduce((ready, instance) => ready && instance.status.ready === 'True', true);
    if (btpInstances.length != 3 || !instancesReady) {
      const is = JSON.stringify(btpInstances, null, 2);
      errors.push(`Expected 3 BTP instances ready but found ${btpInstances.length}:\n${is}`);
    }
    const scBindings = await listResources(`/apis/${scGroup}/${scVersion}/${bindings}`);
    if (scBindings.length != 4) {
      errors.push(`Expected 4 Service Catalog bindings but found ${scBindings.length}`);
    }
    const scInstances = await listResources(`/apis/${scGroup}/${scVersion}/${instances}`);
    if (scInstances.length != 4) {
      errors.push(`Expected 4 Service Catalog instances but found ${scInstances.length}`);
    }
    if (errors.length != 0) {
      await sleep(1000);
    } else {
      break;
    }
  }
  if (errors.length != 0) {
    const info = JSON.stringify(errors, null, 2);
    throw new Error(`Failed to observe migrated ServiceInstances and ServiceBindings.\n${info}`);
  }
}

async function deleteBTPResources() {
  const group = 'services.cloud.sap.com';
  const version = 'v1alpha1';
  const instances = 'serviceinstances';
  const bindings = 'servicebindings';

  let needsPoll = [];
  for (let i = 0; i < 90; i++) { // 15 minutes
    needsPoll = [];
    const k8sBindings = await listResources(`/apis/${group}/${version}/${bindings}`);
    if (k8sBindings.length > 0) {
      needsPoll.push(k8sBindings);
    }
    await deleteK8sObjects(k8sBindings);
    const k8sInstances = await listResources(`/apis/${group}/${version}/${instances}`);
    if (k8sInstances.length > 0) {
      needsPoll.push(k8sInstances);
    }
    await deleteK8sObjects(k8sInstances);
    if (needsPoll.length != 0) {
      await sleep(10000); // 10 seconds
    } else {
      break;
    }
  }
  if (needsPoll.length != 0) {
    const info = JSON.stringify(needsPoll, null, 2);
    throw new Error(`Failed to delete BTP Operator ServiceInstances and ServiceBindings in 15 minutes.\n${info}`);
  }
}

module.exports = {
  provisionPlatform,
  smInstanceBinding,
  cleanupInstanceBinding,
  saveKubeconfig,
  markForMigration,
  readClusterID,
  SMCreds,
  checkPodPresetEnvInjected,
  restartFunctionsPods,
  deleteBTPResources,
  checkMigratedBTPResources,
};
