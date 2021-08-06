var assert = require("assert");
const { k8sCustomApi } = require("../utils");
const {
  getPrometheusRuleGroups,
  prometheusPortForward,
} = require("../monitoring/client");

async function getK8sPrometheusRules() {
  let res = await k8sCustomApi.listNamespacedCustomObject(
    "monitoring.coreos.com",
    "v1",
    "kyma-system",
    "prometheusrules"
  );
  return res.body.items;
}

async function getK8sPrometheusRuleNames() {
  let rules = await getK8sPrometheusRules();
  return rules.map((o) => o.metadata.name);
}

// prometheusPortForward needs to be called before
async function getRegisteredPrometheusRuleNames() {
  let rules = await getPrometheusRuleGroups();
  return rules.map((o) => o.name);
}

function removeNamePrefixes(ruleNames) {
  return ruleNames.map((rule) =>
    rule.replace("monitoring-", "").replace("kyma-", "")
  );
}

async function getNotRegisteredPrometheusRuleNames() {
  let registeredRules = await getRegisteredPrometheusRuleNames();
  let k8sRuleNames = await getK8sPrometheusRuleNames();
  k8sRuleNames = removeNamePrefixes();
  return k8sRuleNames.filter((rule) => !registeredRules.includes(rule));
}

module.exports = { getNotRegisteredPrometheusRuleNames };
