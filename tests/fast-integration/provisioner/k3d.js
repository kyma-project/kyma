const execa = require("execa");
const { debug } = require("../utils");
const { join } = require("path");

async function dockerNetworkCreate() {
  try {
    const { stdout } = await execa("docker", [
      "network",
      "create",
      "kyma",
    ])
    debug(stdout);
    return stdout;
  } catch (err) {
    debug(err.stderr);
    return err;
  }
}
async function dockerNetworkRemove() {
  try {
    const { stdout } = await execa("docker", [
      "network",
      "rm",
      "kyma",
    ])
    debug(stdout);
    return stdout;
  } catch (err) {
    debug(err.stderr);
    return err;
  }
}

async function deleteRegistry() {
  try {
    const { stdout } = await execa("docker", [
      "rm", "-f", "registry.localhost"
    ]);
    debug(stdout);
    return stdout;
  } catch (err) {
    debug(err.stderr);
  }
}

async function createRegistry() {
  const registryDir = join(process.cwd(), 'tmp--registry');
  try {
    const { stdout } = await execa("docker", [
      "run", "-d",
      "-p", "5000:5000",
      "--restart=always",
      "--name", "registry.localhost",
      "--network", "kyma",
      "-v", `${registryDir}:/var/lib/registry`,
      "eu.gcr.io/kyma-project/test-infra/docker-registry-2:20200202"
    ]);
    debug(stdout);
    return stdout;
  } catch (err) {
    debug(err.stderr);
  }
}

async function k3dClusterCreate() {
  const registriesYaml = join(__dirname, 'registries.yaml');

  try {
    const { stdout } = await execa("k3d", [
      "cluster", "create", "kyma",
      "--image", "docker.io/rancher/k3s:v1.19.7-k3s1",
      "--port", "80:80@loadbalancer",
      "--port", "443:443@loadbalancer",
      "--k3s-server-arg", "--no-deploy",
      "--k3s-server-arg", "traefik",
      "--network", "kyma",
      "--volume", `${registriesYaml}:/etc/rancher/k3s/registries.yaml`,
      "--wait",
      "--kubeconfig-switch-context",
      "--timeout", "60s"
    ]);
    debug(stdout);
    return stdout;

  } catch (err) {
    debug(err.stderr);
  }

}
async function k3dClusterDelete() {

  try {
    const { stdout } = await execa("k3d", [
      "cluster", "delete", "kyma",
    ]);
    debug(stdout);
    return stdout;

  } catch (err) {
    debug(err.stderr);
  }

}

async function provisionK3d() {
  await dockerNetworkCreate();
  await createRegistry();
  return await k3dClusterCreate();
}
async function deprovisionK3d() {
  await k3dClusterDelete();
  await deleteRegistry();
  return await dockerNetworkRemove();

}


module.exports = { provisionK3d, deprovisionK3d }