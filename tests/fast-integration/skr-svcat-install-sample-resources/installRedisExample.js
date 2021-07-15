const { join } = require("path");
const k8s = require("@kubernetes/client-node");
const {
    debug,
    k8sCoreV1Api,
    k8sDynamicApi,
    kubectlDelete,
    kubectlApplyDir,
    kubectlApply,
    k8sDelete,
    deleteAllK8sResources,
    deleteNamespaces,
    getAllCRDs,
    k8sApply
  } = require("../utils");


async function installRedisExample(options) {
    options = options || {};

    files = [
      "01_clusteraddonconfiguration.yaml",
      "02_serviceinstance.yaml",
      "03_servicebinding.yaml"
    ]
    const clusteraddonconfigurationPath = options.resourcesPath || join(__dirname, "fixtures", files[0]);
    const serviceinstancePath = options.resourcesPath || join(__dirname, "fixtures", files[1]);
    const servicebindingPath = options.resourcesPath || join(__dirname, "fixtures", files[2]);

    console.log(clusteraddonconfigurationPath)
    console.log(serviceinstancePath)
    console.log(servicebindingPath)

    try {
      await kubectlApply(clusteraddonconfigurationPath);
      console.log(`kubectl apply -f ${clusteraddonconfigurationPath}`)
    } catch (err) {
      throw new Error(`Failed to apply ressource ${files[0]}: ${clusteraddonconfigurationPath}`)
    }
    // TODO: add watch to wait until clusteraddons ready
    try {
      await kubectlApply(serviceinstancePath);
      console.log(`kubectl apply -f ${serviceinstancePath}`)
    } catch (err) {
      throw new Error(`Failed to apply ressource ${files[1]}: ${serviceinstancePath}`)
    }
    // TODO: add watch to wait until serviceinstance ready
    try {
      await kubectlApply(servicebindingPath);
      console.log(`kubectl apply -f ${servicebindingPath}`)
    } catch (err) {
      throw new Error(`Failed to apply ressource ${files[2]}: ${servicebindingPath}`)
    }
}

installRedisExample({})