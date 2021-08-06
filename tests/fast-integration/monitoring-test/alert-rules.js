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

async function getRegisteredPrometheusRules() {
  // prometheusPortForward();
  return await getPrometheusRuleGroups();
}

async function getRegisteredPrometheusRuleNames() {
  let rules = await getRegisteredPrometheusRules();
  return rules.map((o) => o.name);
}

async function getNotRegisteredPrometheusRuleNames() {
  let registeredRules = await getRegisteredPrometheusRuleNames();
  let k8sRuleNames = await getK8sPrometheusRuleNames();
  return k8sRuleNames.filter((rule) => !registeredRules.includes(rule));
}

module.exports = { getNotRegisteredPrometheusRuleNames };
