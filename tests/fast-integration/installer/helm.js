const execa = require("execa");
const fs = require('fs');
const { join } = require("path");
const {debug} = require("../utils");
const notDeployed = "not-deployed";

async function helmStatus(release, namespace) {
  try {
    const { stdout } = await execa("helm", [
      "status",
      release,
      "-n",
      namespace,
      "-ojson",
    ]);
    return JSON.parse(stdout).info.status;
  } catch (err) {
    return notDeployed;
  }
}

async function helmList() {
  const { stdout } = await execa("helm", [
    "ls",
    "-A",
    "-ojson",
  ]);
  return JSON.parse(stdout);
}

async function helmUninstall(release, namespace) {
  debug(`Uninstalling ${release} (${namespace})`);
  const result = await execa("helm", ["uninstall", release, "-n", namespace]);
  debug(`Release ${release} (${namespace}) uninstalled`);
  return result;
}

async function helmInstallUpgrade(release, chart, namespace, values, profile) {
  const args = [
    "upgrade",
    "--wait",
    "-i",
    "-n",
    namespace,
    release,
    chart,
  ];

  if (!!profile) {
    try {
      const profilePath = join(chart, `profile-${profile}.yaml`);
      if (fs.existsSync(profilePath)) {
        args.push("-f", profilePath);
      }
    } catch (err) {
      console.error(`profile-${profile}.yaml file not found in ${chart} - switching to default profile instead`)
    }
  }

  if (!!values) {
    args.push("--set", values);
  }

  await execa("helm", args);
}

async function helmTemplate(release, chart, namespace, values, profile) {
  const args = [
    "template",
    "-n",
    namespace,
    release,
    chart,
  ];

  if (!!profile) {
    try {
      const profilePath = join(chart, `profile-${profile}.yaml`);
      if (fs.existsSync(profilePath)) {
        args.push("-f", profilePath);
      }
    } catch (err) {
      console.error(`profile-${profile}.yaml file not found in ${chart} - switching to default profile instead`)
    }
  }

  if (!!values) {
    args.push("--set", values);
  }

  return execa("helm", args);
}

module.exports = {
  helmInstallUpgrade,
  helmList,
  helmStatus,
  helmUninstall,
  helmTemplate
}