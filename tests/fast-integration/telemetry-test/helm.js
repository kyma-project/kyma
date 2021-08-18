const { exec } = require("child_process");
const { debug } = require("../utils");

function execShellCommand(cmd) {
  const exec = require("child_process").exec;
  return new Promise((resolve, reject) => {
    exec(cmd, (error, stdout, stderr) => {
      if (error) {
        console.warn(error);
      } else {
        debug(`stdout: ${stdout}`);
      }
      resolve(stdout ? stdout : stderr);
    });
  });
}

function installChart(name, path, namespace) {
  return execShellCommand(`helm install -n ${namespace} ${name} ${path}`);
}
function uninstallChart(name, namespace) {
  return execShellCommand(`helm uninstall -n ${namespace} ${name}`);
}

async function InstallAndUninstall() {
  await installChart("mockserver", "./helm/mockserver", "mockserver");
  await uninstallChart("mockserver", "mockserver");
}

// InstallAndUninstall();
// asyncChartInstall("mockserver", "./helm/mockserver", "mockserver");
// asyncChartUninstall("mockserver", "mockserver");

module.exports = {
  installChart,
  uninstallChart,
};
