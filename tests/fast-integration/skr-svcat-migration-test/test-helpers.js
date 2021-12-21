const execa = require('execa');
const fs = require('fs');
const os = require('os');
const { join } = require('path');
const {
  getEnvOrThrow,
  getConfigMap,
  kubectlExecInPod,
  listPods,
  deleteK8sPod,
  sleep,
  deleteAllK8sResources,
  listResources,
} = require('../utils');

class SMCreds {
  static fromEnv() {
    return new SMCreds(
      // TODO: rename to BTP_SM_ADMIN_CLIENTID
      getEnvOrThrow('BTP_OPERATOR_CLIENTID'),
      // TODO: rename to BTP_SM_ADMIN_CLIENTID
      getEnvOrThrow('BTP_OPERATOR_CLIENTSECRET'),
      // TODO: rename to BTP_SM_URL
      getEnvOrThrow('BTP_OPERATOR_URL')
    );
  }

  constructor(clientid, clientsecret, url) {
    this.clientid = clientid;
    this.clientsecret = clientsecret;
    this.url = url;
  }
}

const functions = [
  { name: "svcat-auditlog-api-1", checkEnvVars: "uaa url vendor" },
  { name: "svcat-auditlog-management-1", checkEnvVars: "uaa url vendor" },
  { name: "svcat-html5-apps-repo-1", checkEnvVars: "grant_type saasregistryenabled sap.cloud.service uaa uri vendor" },
];

async function saveKubeconfig(kubeconfig) {
  fs.mkdirSync(`${os.homedir()}/.kube`, true);
  fs.writeFileSync(`${os.homedir()}/.kube/config`, kubeconfig);
}

async function readClusterID() {
  let cm = await getConfigMap("cluster-info", "kyma-system")
  return cm.data.id
}

async function installBTPOperatorHelmChart(creds, clusterId) {
  const btpChart = "https://github.com/kyma-incubator/sap-btp-service-operator/releases/download/v0.1.18-custom/sap-btp-operator-0.1.18.tar.gz";
  const btp = "sap-btp-operator";
  const btpValues = `manager.secret.clientid=${creds.clientId},manager.secret.clientsecret=${creds.clientSecret},manager.secret.url=${creds.smURL},manager.secret.tokenurl=${creds.url},cluster.id=${clusterId}`
  try {
    await helmInstallUpgrade(btp, btpChart, btp, btpValues, null, ["--create-namespace"]);
  } catch (error) {
    if (error.stderr === undefined) {
      throw new Error(`failed to install ${btp}: ${error}`);
    }
    throw new Error(`failed to install ${btp}: ${error.stderr}`);
  }
}

async function installBTPServiceOperatorMigrationHelmChart() {
  const chart = "https://github.com/kyma-incubator/sc-removal/releases/download/0.5.0/sap-btp-operator-migration-v0.5.0.tgz";
  const btp = "sap-btp-service-operator-migration";

  try {
    await helmInstallUpgrade(btp, chart, "sap-btp-operator", null, null, ["--create-namespace"]);
  } catch (error) {
    if (error.stderr === undefined) {
      throw new Error(`failed to install ${btp}: ${error}`);
    }
    throw new Error(`failed to install ${btp}: ${error.stderr}`);
  }
}

async function getFunctionPod(functionName) {
  let labelSelector = `serverless.kyma-project.io/function-name=${functionName},serverless.kyma-project.io/resource=deployment`;
  let res = {};
  for (let i = 0; i < 30; i++) {
    res = await listPods(labelSelector);
    if (res.body.items.length == 1) {
      let pod = res.body.items[0];
      if (pod.status.phase == "Running") {
        return pod
      }
    }
    sleep(10000);
  }
  if (res.body.items.length != 1) {
    let podNames = res.body.items.map(p => p.metadata.name);
    let phases = res.body.items.map(p => p.status.phase);
    throw new Error(`Failed to find function ${functionName} pod in 5 minutes. Expected 1 ${labelSelector} pod with phase "Running" but found ${res.body.items.length}, ${podNames}, ${phases}`);
  }
}

async function checkPodPresetEnvInjected() {
  let cmd = `for v in {vars}; do x="$(eval echo \\$$v)"; if [[ -z "$x" ]]; then echo missing $v env variable; exit 1; else echo found $v env variable; fi; done`;
  for (let f of functions) {
    let pod = await getFunctionPod(f.name);
    let envCmd = cmd.replace("{vars}", f.checkEnvVars);
    await kubectlExecInPod(pod.metadata.name, "function", ["sh", "-c", envCmd]);
  }
}

async function restartFunctionsPods() {
  let podNames = {}
  for (let f of functions) {
    let pod = await getFunctionPod(f.name);
    console.log("delete pod", pod.metadata.name);
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
    for (let f of functions) {
      let labelSelector = `serverless.kyma-project.io/function-name=${f.name},serverless.kyma-project.io/resource=deployment`;
      console.log(`polling pods with labelSelector ${labelSelector}`);
      let res = {};
      try {
        res = await listPods(labelSelector);
      } catch (err) {
        throw new Error(`failed to list pods with labelSelector ${labelSelector}: ${err}`);
      }
      if (res.body.items.length != 1) {
        // there are either multiple or 0 pods for the function, we need to wait
        let pn = res.body.items.map(p => { return { "pod name": p.metadata.name, phase: p.status.phase } });
        needsPoll.push({ "function name": f.name, pods: pn });
        continue;
      }
      let pod = res.body.items[0]
      let pn = pod.metadata.name
      let psp = pod.status.phase
      if (pn == podNames[f.name]) {
        // there is single pod for the function, but it is still the old one
        needsPoll.push({ "function name": f.name, pods: [{ "pod name": pn, phase: psp }] });
        continue;
      }
      if (psp != "Running") {
        // there is single pod for the function, it has new name, but it's not in Running state
        needsPoll.push({ "function name": f.name, pods: [{ "pod name": pn, phase: psp }] });
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
    let info = JSON.stringify(needsPoll, null, 2);
    let originalNames = JSON.stringify(podNames, null, 2);
    throw new Error(`Failed to restart function pods in 100 seconds. Expecting exactly one pod for each function with new unique names and in ready status but found:\n${info}\n\nPod names before restart:\n${originalNames}`);
  }
  console.log("functions pods successfully restarted")
}

async function provisionPlatform(creds, svcatPlatform) {
  let args = [];
  try {
    args = [`login`, `-a`, creds.url, `--param`, `subdomain=e2etestingscmigration`, `--auth-flow`, `client-credentials`]
    await execa(`smctl`, args.concat([`--client-id`, creds.clientid, `--client-secret`, creds.clientsecret]));

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
    args = [`register-platform`, svcatPlatform, `kubernetes`, `-o`, `json`]
    let registerPlatformOut = await execa(`smctl`, args);
    let platform = JSON.parse(registerPlatformOut.stdout)

    return {
      clusterId: platform.id,
      name: platform.name,
      credentials: platform.credentials.basic,
    }

  } catch (error) {
    if (error.stderr === undefined) {
      throw new Error(`failed to process output of "smctl ${args.join(' ')}": ${error}`);
    }
    throw new Error(`failed "smctl ${args.join(' ')}": ${error.stderr}`);
  }
}

async function smInstanceBinding(btpOperatorInstance, btpOperatorBinding) {
  let args = [];
  try {
    args = [`provision`, btpOperatorInstance, `service-manager`, `service-operator-access`, `--mode=sync`]
    await execa(`smctl`, args);

    // Move to Operator Install
    args = [`bind`, btpOperatorInstance, btpOperatorBinding, `--mode=sync`];
    await execa(`smctl`, args);
    args = [`get-binding`, btpOperatorBinding, `-o`, `json`];
    let out = await execa(`smctl`, args);
    let b = JSON.parse(out.stdout)
    let c = b.items[0].credentials

    return {
      clientId: c.clientid,
      clientSecret: c.clientsecret,
      smURL: c.sm_url,
      url: c.url,
      instanceId: b.items[0].service_instance_id,
    }

  } catch (error) {
    if (error.stderr === undefined) {
      throw new Error(`failed to process output of "smctl ${args.join(' ')}": ${error}`);
    }
    throw new Error(`failed "smctl ${args.join(' ')}": ${error.stderr}`);
  }
}

async function markForMigration(creds, svcatPlatform, btpOperatorInstanceId) {
  let errors = [];
  let args = [];
  try {
    args = [`login`, `-a`, creds.url, `--param`, `subdomain=e2etestingscmigration`, `--auth-flow`, `client-credentials`]
    await execa(`smctl`, args.concat([`--client-id`, creds.clientid, `--client-secret`, creds.clientsecret]));
  } catch (error) {
    errors = errors.concat([`failed "smctl ${args.join(' ')}": ${error.stderr}\n${error}`]);
  }

  try {
    // usage: smctl curl -X PUT -d '{"sourcePlatformID": ":platformID"}' /v1/migrate/service_operator/:instanceID
    let data = { sourcePlatformID: svcatPlatform }
    args = ['curl', '-X', 'PUT', '-d', JSON.stringify(data), '/v1/migrate/service_operator/' + btpOperatorInstanceId]
    await execa('smctl', args)
  } catch (error) {
    errors = errors.concat([`failed "smctl ${args.join(' ')}": ${error.stderr}\n${error}`]);
  }
  if (errors.length > 0) {
    throw new Error(errors.join(", "));
  }
}

async function cleanupInstanceBinding(creds, svcatPlatform, btpOperatorInstance, btpOperatorBinding) {
  let errors = [];
  let args = [];
  try {
    args = [`login`, `-a`, creds.url, `--param`, `subdomain=e2etestingscmigration`, `--auth-flow`, `client-credentials`]
    await execa(`smctl`, args.concat([`--client-id`, creds.clientid, `--client-secret`, creds.clientsecret]));
  } catch (error) {
    errors = errors.concat([`failed "smctl ${args.join(' ')}": ${error.stderr}\n${error}`]);
  }

  try {
    args = [`unbind`, btpOperatorInstance, btpOperatorBinding, `-f`, `--mode=sync`];
    let { stdout } = await execa(`smctl`, args);
    if (stdout !== "Service Binding successfully deleted.") {
      errors = errors.concat([`failed "smctl ${args.join(' ')}": ${stdout}`])
    }
  } catch (error) {
    errors = errors.concat([`failed "smctl ${args.join(' ')}": ${error.stderr}\n${error}`]);
  }

  try {
    // hint: probably should fail cause that instance created other instannces (after the migration is done)
    args = [`deprovision`, btpOperatorInstance, `-f`, `--mode=sync`];
    let { stdout } = await execa(`smctl`, args);
    if (stdout !== "Service Instance successfully deleted.") {
      errors = errors.concat([`failed "smctl ${args.join(' ')}": ${stdout}`])
    }
  } catch (error) {
    errors = errors.concat([`failed "smctl ${args.join(' ')}": ${error.stderr}\n${error}`]);
  }

  try {
    args = [`delete-platform`, svcatPlatform, `-f`, "--cascade"];
    await execa(`smctl`, args);
    // if (stdout !== "Platform(s) successfully deleted.") {
    //     errors = errors.concat([`failed "smctl ${args.join(' ')}": ${stdout}`])
    // }
  } catch (error) {
    errors = errors.concat([`failed "smctl ${args.join(' ')}": ${error.stderr}\n${error}`]);
  }

  if (errors.length > 0) {
    throw new Error(errors.join(", "));
  }
}

async function deleteBTPResources() {
  const group = "services.cloud.sap.com";
  const version = "v1alpha1";
  const instances = "serviceinstances";
  const bindings = "servicebindings";
  const keepFinalizers = true;
  await deleteAllK8sResources(`/apis/${group}/${version}/${instances}`, {}, 10, 1000, keepFinalizers);
  await deleteAllK8sResources(`/apis/${group}/${version}/${bindings}`, {}, 10, 1000, keepFinalizers);

  let needsPoll = [];
  for (let i = 0; i < 90; i++) { // 15 minutes
    needsPoll = [];
    let k8sInstances = listResources(`/apis/${group}/${version}/${instances}`);
    if (k8sInstances > 1) {
      needsPoll.push(k8sInstances);
    }
    let k8sBindings = listResources(`/apis/${group}/${version}/${bindings}`);
    if (k8sBindings > 1) {
      needsPoll.push(k8sBindings);
    }
    if (needsPoll.length != 0) {
      await sleep(10000); // 10 seconds
    } else {
      break;
    }
  }
  if (needsPoll.length != 0) {
    let info = JSON.stringify(needsPoll, null, 2);
    throw new Error(`Failed to delete BTP Operator ServiceInstances and ServiceBindings in 15 minutes.\n${info}`);
  }
}

async function helmInstallUpgrade(release, chart, namespace, values, profile, additionalArgs) {
  const args = [
    'upgrade',
    '--wait',
    '-i',
    '-n',
    namespace,
    release,
    chart,
  ];

  if (Array.isArray(additionalArgs)) {
    args.push(...additionalArgs);
  }

  if (!!profile) {
    try {
      const profilePath = join(chart, `profile-${profile}.yaml`);
      if (fs.existsSync(profilePath)) {
        args.push('-f', profilePath);
      }
    } catch (err) {
      console.error(`profile-${profile}.yaml file not found in ${chart} - switching to default profile instead`)
    }
  }

  if (!!values) {
    args.push('--set', values);
  }

  await execa('helm', args);
}

module.exports = {
  provisionPlatform,
  smInstanceBinding,
  cleanupInstanceBinding,
  installBTPOperatorHelmChart,
  installBTPServiceOperatorMigrationHelmChart,
  saveKubeconfig,
  markForMigration,
  readClusterID,
  SMCreds,
  checkPodPresetEnvInjected,
  restartFunctionsPods,
  deleteBTPResources,
};