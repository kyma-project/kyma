const { join } = require("path");
const {
    kubectlApply,
    kubectlDeleteDir,
    waitForPodWithLabel,
    waitForFunction,
    waitForServiceInstance,
    waitForServiceBinding,
    waitForServiceBindingUsage,
    waitForClusterAddonsConfiguration,
    getSecrets,
    getPodPresets,
    debug,
  } = require("../utils");
const {
  performance,
} = require('perf_hooks');

async function installResource(manifest, resource, name, namespace) {

  try {
    debug(`Applying ${resource} manifest ${manifest}`)
    await kubectlApply(manifest);
  } catch (error) {
    throw new Error(`Failed to apply resource ${name}: ${manifest}.`)
  }

  try {
    switch (resource) {
      case "serviceInstance":
        debug(`Waiting for ${resource} ${name}`)
        await waitForServiceInstance(name, namespace)
        break;
      case "function":
        debug(`Waiting for ${resource} ${name}`)
        await waitForFunction(name, namespace)
        break;
      case "servicebinding":
        debug(`Waiting for ${resource} ${name}`)
        await waitForServiceBinding(name, namespace)
        break;
      case "servicebindingusage":
        debug(`Waiting for ${resource} ${name}`)
        await waitForServiceBindingUsage(name, namespace)
        break;
      default:
        debug("Not waiting for resource installation.")
        break;
    }
  } catch (error) {
    console.log(error)
    throw new Error(`Failed to wait for resource ${resource} with name ${name}.`)
  }

}

async function installRedisExample(options) {
  // console.time('installRedisExample')
  let t0 = performance.now()

  options = options || {};
  
  debug(`Waiting for pod with label app=helm-broker...`)
  await waitForPodWithLabel("app", "helm-broker", "kyma-system");

  const clusteraddonconfigurationPath = options.resourcesPath || join(__dirname, "fixtures", "01_clusteraddonconfiguration.yaml");
  debug(`Applying manifest ${clusteraddonconfigurationPath}`)
  await kubectlApply(clusteraddonconfigurationPath);
  await waitForClusterAddonsConfiguration("redis-addon")
  
  // Deploying redis instance with function and servicebindings
  const serviceInstanceManifestPath = options.resourcesPath || join(__dirname, "fixtures", "02_serviceinstance_redis.yaml");
  const functionManifestPath = options.resourcesPath || join(__dirname, "fixtures", "03_function_redis.yaml");
  const funcSBManifestPath = options.resourcesPath || join(__dirname, "fixtures", "04_func-sb_redis.yaml");
  const instSBManifestPath = options.resourcesPath || join(__dirname, "fixtures", "04_inst-sb_redis.yaml");
  const sbuManifestPath = options.resourcesPath || join(__dirname, "fixtures", "05_sbu_redis.yaml");
  
  await installResource(serviceInstanceManifestPath, "serviceInstance", "hb-instbind-redis-1", "default")
  await installResource(functionManifestPath, "function", "hb-instbind-redis-1", "default")
  await installResource(funcSBManifestPath, "servicebinding", "func-sb-redis-function-1", "default")
  await installResource(instSBManifestPath, "servicebinding", "inst-sb-redis-function-1", "default")
  await installResource(sbuManifestPath, "servicebindingusage", "hb-instbind-redis-1", "default")
  
  // return console.timeEnd('installRedisExample')
  let t1 = performance.now()
  console.log(`Finished deployment of Redis in ${t1-t0} ms`)
  return {resource: "Redis", duration: t1-t0}
}

async function installAuditlogExample(options) {
  // console.time('installAuditlogExample')
  let t0 = performance.now()
  
  options = options || {};
  
  const serviceInstanceManifestPath = options.resourcesPath || join(__dirname, "fixtures", "02_serviceinstance_auditlog.yaml");
  const functionManifestPath = options.resourcesPath || join(__dirname, "fixtures", "03_function_auditlog.yaml");
  const funcSBManifestPath = options.resourcesPath || join(__dirname, "fixtures", "04_func-sb_auditlog.yaml");
  const instSBManifestPath = options.resourcesPath || join(__dirname, "fixtures", "04_inst-sb_auditlog.yaml");
  const sbuManifestPath = options.resourcesPath || join(__dirname, "fixtures", "05_sbu_auditlog.yaml");
  
  await installResource(serviceInstanceManifestPath, "serviceInstance", "svcat-auditlog-api-1", "default")
  await installResource(functionManifestPath, "function", "svcat-auditlog-api-1", "default")
  await installResource(funcSBManifestPath, "servicebinding", "func-sb-svcat-auditlog-api-1", "default")
  await installResource(instSBManifestPath, "servicebinding", "inst-sb-svcat-auditlog-api-1", "default")
  await installResource(sbuManifestPath, "servicebindingusage", "func-sbu-svcat-auditlog-api-1", "default")
  
  // return console.timeEnd('installAuditlogExample')
  let t1 = performance.now()
  console.log(`Finished deployment of Audit-Log in ${t1-t0} ms`)
  return {resource: "Audit-Log", duration: t1-t0}
}

async function installHTML5AppsRepoExample(options) {
  // console.time('installHTML5AppsRepoExample')
  let t0 = performance.now()

  options = options || {};
  
  const serviceInstanceManifestPath = options.resourcesPath || join(__dirname, "fixtures", "02_serviceinstance_html5appsrepo.yaml");
  const functionManifestPath = options.resourcesPath || join(__dirname, "fixtures", "03_function_html5appsrepo.yaml");
  const funcSBManifestPath = options.resourcesPath || join(__dirname, "fixtures", "04_func-sb_html5appsrepo.yaml");
  const instSBManifestPath = options.resourcesPath || join(__dirname, "fixtures", "04_inst-sb_html5appsrepo.yaml");
  const sbuManifestPath = options.resourcesPath || join(__dirname, "fixtures", "05_sbu_html5appsrepo.yaml");
  
  await installResource(serviceInstanceManifestPath, "serviceInstance", "svcat-html5-apps-repo-1", "default")
  await installResource(functionManifestPath, "function", "svcat-html5-apps-repo-1", "default")
  await installResource(funcSBManifestPath, "servicebinding", "func-sb-svcat-html5-apps-repo-1", "default")
  await installResource(instSBManifestPath, "servicebinding", "inst-sb-svcat-html5-apps-repo-1", "default")
  await installResource(sbuManifestPath, "servicebindingusage", "func-sbu-svcat-html5-apps-repo-1", "default")
  
  // return console.timeEnd('installHTML5AppsRepoExample')
  let t1 = performance.now()
  console.log(`Finished deployment of HTML5-Apps-Repo in ${t1-t0} ms`)
  return {resource: "HTML5-Apps-Repo", duration: t1-t0}
}

async function installAuditManagementExample(options) {
  // console.time('installAuditManagementExample')
  let t0 = performance.now()

  options = options || {};
  
  const serviceInstanceManifestPath = options.resourcesPath || join(__dirname, "fixtures", "02_serviceinstance_auditlogmanagement.yaml");
  const functionManifestPath = options.resourcesPath || join(__dirname, "fixtures", "03_function_auditlogmanagement.yaml");
  const funcSBManifestPath = options.resourcesPath || join(__dirname, "fixtures", "04_func-sb_auditlogmanagement.yaml");
  const instSBManifestPath = options.resourcesPath || join(__dirname, "fixtures", "04_inst-sb_auditlogmanagement.yaml");
  const sbuManifestPath = options.resourcesPath || join(__dirname, "fixtures", "05_sbu_auditlogmanagement.yaml");
  
  await installResource(serviceInstanceManifestPath, "serviceInstance", "svcat-auditlog-management-1", "default")
  await installResource(functionManifestPath, "function", "svcat-auditlog-management-1", "default")
  await installResource(funcSBManifestPath, "servicebinding", "func-sb-svcat-auditlog-management-1", "default")
  await installResource(instSBManifestPath, "servicebinding", "inst-sb-svcat-auditlog-management-1", "default")
  await installResource(sbuManifestPath, "servicebindingusage", "func-sbu-svcat-auditlog-management-1", "default")
  
  // return console.timeEnd('installAuditManagementExample')
  let t1 = performance.now()
  console.log(`Finished deployment of Audit-Management in ${t1-t0} ms`)
  return {resource: "Audit-Management", duration: t1-t0}
}

async function destroy() {  
  try {
    let path = join(__dirname, "fixtures")
    debug(`Destroying all ressources from ${path}`)
    await kubectlDeleteDir(path, "default")
  } catch (err) {
    throw new Error(`Error destroying ressources from ${path}`)
  }
}

function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

async function goodNight() {
  console.log('Taking a break...');
  await sleep(1200000); //20min
  console.log('Waking up...');
}

async function storeSecretsAndPresets() {
  let secretsResp = await getSecrets("default")
  let podPresetsResp = await getPodPresets("default")

  let secrets = []
  let podPresets = []

  secretsResp.forEach(function (secret) {
    secrets.push(secret.metadata.name)
  })

  podPresetsResp.forEach(function (podPreset) {
    podPresets.push(podPreset.metadata.name)
  })

  return { secrets, podPresets }
}

async function checkSecrets(existing) {
  let allSecrets = existing

  // this are fixed values from resource names of test fixtures
  let reference = [
    "hb-redis-micro",
    "sh.helm.release.v1.hb-redis-micro",
    "func-sb-svcat-auditlog-management-1",
    "func-sb-svcat-auditlog-api-1",
    "inst-sb-svcat-auditlog-api-1",
    "inst-sb-svcat-auditlog-management-1",
    "func-sb-svcat-html5-apps-repo-1",
    "inst-sb-svcat-html5-apps-repo-1",
    "func-sb-redis-function-1",
    "inst-sb-redis-function-1",
  ]

  let missing = reference

  // iterate through the list of all existing secrets in namespace
  // and check if this list contains our well known secrets
  allSecrets.forEach( function (secretName) {
    for (var i=0; i < reference.length; i++) {
      if (secretName.includes(reference[i])) {
        let index = missing.indexOf(reference[i])
        if (index !== -1) {
          missing.splice(index, 1)
        }
      }
    }
  })

  // if missing > 0 -> there are some secrets lost
  debug(`Missing ${missing.length} secrets`)
}

async function checkPodPresets(expected, existing) {
  let allPodPresets = existing

  let missing = expected
  debug("checking if all podpresets still exists: " + missing)

  // iterate through the list of all existing PodPresets in namespace
  // and check if this list still contains the same PodPresets
  allPodPresets.forEach( function (podPresetName) {
    for (var i=0; i < expected.length; i++) {
      if (podPresetName.includes(expected[i])) {
        let index = missing.indexOf(expected[i])
        if (index !== -1) {
          missing.splice(index, 1)
        }
      }
    }
  })

  // if missing > 0 -> there are some secrets lost
  debug(`Missing ${missing.length} PodPresets`)
}

async function deploy() {
  let times = []

  await waitForPodWithLabel("app", "service-catalog-addons-service-binding-usage-controller", "kyma-system");
  await waitForPodWithLabel("app", "service-catalog-ui", "kyma-system");
  await waitForPodWithLabel("app", "service-catalog-catalog-controller-manager", "kyma-system");
  await waitForPodWithLabel("app", "service-catalog-catalog-webhook", "kyma-system");
  await waitForPodWithLabel("app", "service-broker-proxy-k8s", "kyma-system");

  times.push(installRedisExample())
  times.push(installAuditlogExample())
  times.push(installHTML5AppsRepoExample())
  times.push(installAuditManagementExample())
  
  return Promise.all(times).then(function () {
    console.log(`\nSuccessfully deployed all resources:`)
    let items = arguments[0]
    items.sort(function(a, b) {
      if (a.duration < b.duration) {
        return -1;
      }
      if (a.duration > b.duration) {
        return 1;
      }
      return 0;
    })
    console.table(items, ["resource", "duration"])
  }, function (error) {
    console.log(error)
  })
}

module.exports = {
  deploy,
  destroy,
  storeSecretsAndPresets,
  checkSecrets,
  checkPodPresets,
  goodNight
};
