const execa = require("execa");
const fs = require('fs');
const { join } = require("path");
const { installIstio, upgradeIstio } = require("./istioctl");
const pRetry = require("p-retry");
const {
  debug,
  k8sCoreV1Api,
  kubectlDeleteDir,
  kubectlDelete,
  kubectlApplyDir,
  kubectlApply,
  k8sDelete,
  deleteAllK8sResources,
  deleteNamespaces,
  getAllCRDs,
  k8sApply
} = require("../utils");
const { map } = require("lodash");
const notDeployed = "not-deployed";
const kymaCrds = require("./kyma-crds");

async function waitForNodesToBeReady(timeout = "180s") {
  await execa("kubectl", [
    "wait",
    "--for=condition=Ready",
    "nodes",
    "--all",
    `--timeout=${timeout}`,
  ]);
  debug("All nodes are ready!");
}


async function updateCoreDNSConfigMap() {

  const cmName = "coredns";
  const cmNamespace = "kube-system";

  const { body } = await k8sCoreV1Api.readNamespacedConfigMap(
    "coredns",
    "kube-system"
  );

  const { stdout: registryIP } = await execa("docker", [
    "inspect",
    "-f",
    "{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}",
    "/registry.localhost",
  ]).catch(() => "127.0.0.1");

  // this file need to be updated when we update k3s to use k8s 1.20
  body.data["Corefile"] = `registry.localhost:53 {
    hosts {
        ${registryIP} registry.localhost
    }
}
.:53 {
    errors
    health
    rewrite name regex (.*)\.local\.kyma\.dev istio-ingressgateway.istio-system.svc.cluster.local
    ready
    kubernetes cluster.local in-addr.arpa ip6.arpa {
      pods insecure
      upstream
      fallthrough in-addr.arpa ip6.arpa
    }
    hosts /etc/coredns/NodeHosts {
      reload 1s
      fallthrough
    }
    prometheus :9153
    forward . /etc/resolv.conf
    cache 30
    loop
    reload
    loadbalance
}
`;
  await k8sCoreV1Api.replaceNamespacedConfigMap(cmName, cmNamespace, body);
}

async function removeKymaGatewayCertsYaml(pathToResources) {
  // TODO: find some other mechanism, deleting that file is destructive
  // we can copy resources directory so that we can act on it
  try {
    await execa("rm", [
      join(
        pathToResources,
        "core/charts/gateway/templates/kyma-gateway-certs.yaml"
      ),
    ]);
  } catch (err) {
    console.log("kyma-gateway-certs.yaml already deleted");
  }
}

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
  await execa("helm", ["uninstall", release, "-n", namespace]);
}

async function helmInstallUpgrade(release, chart, namespace, values, profile) {
  const args = [
    "upgrade",
    "--wait",
    "--force",
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

async function installRelease(
  release,
  namespace = "kyma-system",
  chart,
  values,
  profile,
  isUpgrade
) {

  let status = await helmStatus(release, namespace);
  switch (status) {
    case "pending-install":
      debug(
        `Deleting ${release} from namespace ${namespace} because previous installation got stuck in pending-install`
      );
      await helmUninstall(release, namespace);
      break;
    case "failed":
      debug(
        `Deleting ${release} from namespace ${namespace} because previous installation failed`
      );
      await helmUninstall(release, namespace);
      break;
    case notDeployed:
      debug(`Installing release ${release}`);
      await helmInstallUpgrade(release, chart, namespace, values, profile);
      break;
    case "deployed":
      if (isUpgrade) {
        debug(`Upgrading release ${release}`);
        await helmInstallUpgrade(release, chart, namespace, values, profile);

      } else {
        debug(`Release ${release} already installed - skipped`);
        return;
      }
    default:
      break;
  }

  status = await helmStatus(release, namespace);
  if (status !== "deployed") {
    throw new Error(
      `Release ${release} is in status ${status}, which is not "deployed"`
    );
  }
  debug(`Release ${release} deployed`);
}

async function chartList() {
  const gardernerDomain = await getGardenerDomain();
  const isGardener = process.env["GARDENER"] || (gardernerDomain) ? "true" : "false";
  const domain = process.env["KYMA_DOMAIN"] || gardernerDomain || "local.kyma.dev";

  const overrides = `global.isLocalEnv=false,global.ingress.domainName=${domain},global.environment.gardener=${isGardener},global.domainName=${domain},global.tlsCrt=ZHVtbXkK`;
  // https://github.com/kyma-project/test-infra/pull/2967
  let registryOverrides = `dockerRegistry.enableInternal=false,dockerRegistry.serverAddress=registry.localhost:5000,dockerRegistry.registryAddress=registry.localhost:5000,global.ingress.domainName=${domain},containers.manager.envs.functionBuildExecutorImage.value=eu.gcr.io/kyma-project/external/aerfio/kaniko-executor:v1.3.0`;
  if (isGardener == "true") {
    registryOverrides = `dockerRegistry.enableInternal=true,global.ingress.domainName=${domain}`
  }

  const kymaCharts = [
    {
      release: "pod-preset",
      namespace: "kyma-system",
      customPath: (root) =>
        join(root, "cluster-essentials", "charts", "pod-preset"),
    },
    {
      release: `cluster-users`,
      namespace: "kyma-system",
      values: `${overrides}`,
    },
    {
      release: "core",
      namespace: "kyma-system",
      values: `${overrides}`,
    },
    {
      release: "ory",
      namespace: "kyma-system",
      values: `${overrides}`,
    },
    {
      release: "serverless",
      namespace: "kyma-system",
      values: registryOverrides,
    },
    {
      release: "dex",
      namespace: "kyma-system",
      values: `${overrides},resources.requests.cpu=10m`,
    },
    {
      release: "api-gateway",
      namespace: "kyma-system",
      values: `${overrides},deployment.resources.requests.cpu=10m`,
    },
    {
      release: "rafter",
      namespace: "kyma-system",
      values: `${overrides}`,
    },
    {
      release: "service-catalog",
      namespace: "kyma-system",
      values: `${overrides}`,
    },
    {
      release: "service-catalog-addons",
      namespace: "kyma-system",
      values: `${overrides}`,
    },
    {
      release: "helm-broker",
      namespace: "kyma-system",
      values: `${overrides}`,
    },
    {
      release: "console",
      namespace: "kyma-system",
      values: `${overrides},pamela.enabled=false`,
    },
    {
      release: "knative-eventing",
      namespace: "knative-eventing",
      values: `${overrides}`,
    },
    {
      release: "application-connector",
      namespace: "kyma-integration",
      values: `${overrides}`,
    },
    {
      release: "knative-provisioner-natss",
      namespace: "knative-eventing",
      values: `${overrides}`,
    },
    {
      release: "nats-streaming",
      namespace: "natss",
    },
    {
      release: "event-sources",
      namespace: "kyma-system",
    },
    {
      release: "monitoring",
      namespace: "kyma-system",
      values: `${overrides}`,
    },
    {
      release: "kiali",
      namespace: "kyma-system",
      values: `${overrides}`,
    },
    {
      release: "tracing",
      namespace: "kyma-system",
      values: `${overrides}`,
    },
    {
      release: "logging",
      namespace: "kyma-system",
      values: `${overrides}`,
    },
    {
      release: "ingress-dns-cert",
      namespace: "istio-system",
      values: `global.ingress.domainName=${domain},global.environment.gardener=${isGardener}`,
      customPath: () => join(__dirname, "charts", "ingress-dns-cert"),
    },
  ];
  return kymaCharts;
}

async function getGardenerDomain() {
  try {
    const { body } = await k8sCoreV1Api.readNamespacedConfigMap(
      "shoot-info",
      "kube-system"
    );
    return body.data.domain;
  } catch (err) {
    if (err.statusCode == 404) {
      return null;
    }
    throw err;
  }
}
async function uninstallIstio() {
  const crds = await getAllCRDs();
  const istioCRDs = crds.filter(crd => crd.spec.group.endsWith('istio.io'));
  await k8sDelete(istioCRDs);
  await deleteNamespaces(["istio-system"], true);
}

async function uninstallKyma(options
) {
  const releases = await helmList();

  await Promise.all(releases.map((r) => helmUninstall(r.name, r.namespace).catch())); // ignore errors during uninstall ()
  await kubectlDelete(join(__dirname, "installer-local.yaml")); // needed for the console to start
  if (!options.skipICrd) {
    const crds = await getAllCRDs();
    await k8sDelete(crds.filter(crd => kymaCrds.includes(crd.metadata.name)));
  }
  await kubectlDelete(join(__dirname, "system-namespaces.yaml"));
  await deleteAllK8sResources('/api/v1/namespaces/kyma-system/secrets');
  if (!options.skipIstio) {
    await uninstallIstio();
  }

}

async function installKyma(
  installLocation = join(__dirname, "..", "..", "..", "resources"),
  istioVersion,
  ignoredComponents = "monitoring,tracing,kiali,logging,console,cluster-users,dex",
  isUpgrade = false,
) {
  console.log('Installing Kyma from folder', installLocation);
  await waitForNodesToBeReady();
  const crdsBefore = await getAllCRDs();
  if (isUpgrade) {
    await upgradeIstio(istioVersion);
  } else {
    await updateCoreDNSConfigMap();
    await installIstio(istioVersion);
  }

  await removeKymaGatewayCertsYaml(installLocation);
  await kubectlApply(join(__dirname, "installer-local.yaml")); // needed for the console to start
  await kubectlApply(join(__dirname, "system-namespaces.yaml"));
  await kubectlApplyDir(
    join(installLocation, "cluster-essentials/files"),
    "kyma-system"
  );
  const kymaCharts = await chartList();
  await Promise.all(
    kymaCharts
      .filter((arg) => !ignoredComponents.includes(arg.release))
      .map(({ release, namespace, values, customPath, profile }) => {
        const chartLocation = !!customPath
          ? customPath(installLocation)
          : join(installLocation, release);
        return pRetry(
          async () =>
            installRelease(release, namespace, chartLocation, values, profile || "evaluation", isUpgrade),
          {
            retries: 10,
            onFailedAttempt: async (err) => {
              console.log(`retrying install of ${release}`);
              console.log(err);
            },
          }
        );
      })
  );
  const crdsAfter = await getAllCRDs();
  const installedCrds = crdsAfter.filter(crd => !crdsBefore.some(c => c.metadata.name == crd.metadata.name));
  debug("Installed crds:")
  debug(installedCrds.map(crd => crd.metadata.name));
}

module.exports = {
  installKyma,
  uninstallKyma
};

