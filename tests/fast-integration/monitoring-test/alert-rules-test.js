var assert = require("assert");
const { watch } = require("fs");
const k8s = require("@kubernetes/client-node");
const {
  k8sCoreV1Api,
  k8sAppsApi,
  k8sExtensionApi,
  k8sDynamicApi,
  getAllCRDs,
  getAllResourceTypes,
  listResourceNames,
  listResources,
  k8sCustomApi,
} = require("../utils");
const {
  getPrometheusRuleGroups,
  prometheusPortForward,
} = require("../monitoring/client");

async function getCRDs() {
  let res = await k8sCustomApi.listNamespacedCustomObject(
    "monitoring.coreos.com",
    "v1",
    "kyma-system",
    "prometheusrules"
  );
  let rules = res.body.items;
  let k8sRuleNames = rules.map((o) => o.metadata.name);

  // check Prometheus
  prometheusPortForward();
  let promres = await getPrometheusRuleGroups();
  let registeredRules = promres.map((o) => o.name);
  console.log("Registered rules", registeredRules);
  console.log(
    "Not registered",
    k8sRuleNames.filter((rule) => !registeredRules.includes(rule))
  );
}

getCRDs();
