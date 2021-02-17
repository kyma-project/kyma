const fs = require("fs");
const execa = require("execa");
const { debug } = require("../utils");
const { join } = require("path");

function istioctlLocation(version) {
  return join(__dirname, "..", `istio-${version}`, "bin", "istioctl");
}

function isIstioctlDownloaded(version) {
  try {
    return fs.existsSync(istioctlLocation(version));
  } catch (err) {
    return false;
  }
}

async function upgradeIstio(version) {
  await ensureIstioctl(version);

  const istioctl = istioctlLocation(version);
  const { stdout } = await execa(
    istioctl,
    ["upgrade", "-y", "-f", join(__dirname, "config-istio.yaml")],
    {
      shell: true,
    }
  );

  debug(stdout);
}

async function installIstio(version, istioProfile = "demo") {
  await ensureIstioctl(version);
  const istioctl = istioctlLocation(version);
  const { stdout } = await execa(
    istioctl,
    [
      "install",
      "-y",
      "--set",
      `profile=${istioProfile}`,
      "-f",
      join(__dirname, "config-istio.yaml"),
    ],
    {
      shell: true,
    }
  );

  debug(stdout);
}

async function ensureIstioctl(version) {
  if (!isIstioctlDownloaded(version)) {
    await downloadIstioctl(version);
  }
}

async function downloadIstioctl(version) {
  const { stdout } = await execa.command(
    `curl -sL https://istio.io/downloadIstio | ISTIO_VERSION=${version} sh -`,
    {
      shell: true,
    }
  );

  debug(stdout);
}

module.exports = { installIstio, upgradeIstio };
