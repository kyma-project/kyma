const downloader = require("./chart-downloader");
const {
  helmInstallUpgrade,
  helmStatus,
  helmUninstall,
  helmList,
  helmTemplate
} = require("./helm");
const execa = require("execa");
const { join } = require("path");
const { installIstio, upgradeIstio } = require("./istioctl");
const pRetry = require("p-retry");
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
const notDeployed = "not-deployed";
const kymaCrds = require("./kyma-crds");
const defaultIstioVersion = "1.8.2";


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
function skipHelmTest(r) {
  return !(r.metadata.annotations && r.metadata.annotations["helm.sh/hook"] && r.metadata.annotations["helm.sh/hook"].includes("test-success"));
}
function skipNull(r) {
  return (r!=null);
}

async function applyRelease(
  release,
  namespace = "kyma-system",
  chart,
  values,
  profile,
  isUpgrade
) {
  const { stdout } = await helmTemplate(release, chart, namespace, values, profile);
  const yamls = k8s.loadAllYaml(stdout);

  await k8sApply(yamls.filter(skipNull).filter(skipHelmTest), namespace);
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

async function chartList(options) {
  const gardernerDomain = await getGardenerDomain();
  const isGardener = process.env["GARDENER"] || (gardernerDomain) ? "true" : "false";
  const domain = process.env["KYMA_DOMAIN"] || gardernerDomain || "local.kyma.dev";
  const isCompassEnabled = !!options.withCompass;
  const overrides = `global.isLocalEnv=false,global.ingress.domainName=${domain},global.environment.gardener=${isGardener},global.domainName=${domain},global.tlsCrt=ZHVtbXkK,global.isBEBEnabled=true,global.disableLegacyConnectivity=${isCompassEnabled}`;
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
      release: "eventing",
      namespace: "kyma-system",
      values: `${overrides}`
    },
    {
      release: "application-connector",
      namespace: "kyma-integration",
      values: `${overrides}`,
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
      release: "compass-runtime-agent",
      namespace: "compass-system",
      values: `${overrides}`,
      filter: isCompassEnabled
    },
    {
      release: "ingress-dns-cert",
      namespace: "istio-system",
      values: `global.ingress.domainName=${domain},global.environment.gardener=${isGardener}`,
      customPath: () => join(__dirname, "charts", "ingress-dns-cert"),
    },
  ];

  return kymaCharts.filter(c => c.filter == undefined || c.filter);
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
/**
 * 
 * Uninstalls Kyma
 * @param {Object} options Uninstallation options
 * @param {boolean} options.skipCrd Do not delete CRDs
 * @param {Array<string>} options.skipComponents List of components that should not be uninstalled
 * @param {Array<string>} options.components List of components to uninstall
 * @param {boolean} options.deleteNamespaces
 */
async function uninstallKyma(options) {
  let releases = await helmList();
  debug("Releases in the cluster:", releases);
  const components = options.components;
  const skipComponents = options.skipComponents;
  if (components) {
    releases = releases.filter(r => components.includes(r.name))
  }
  if (skipComponents) {
    releases = releases.filter(r => !skipComponents.includes(r.name))
  }
  debug("Releases to uninstall:", releases);

  await Promise.allSettled(releases.map((r) => helmUninstall(r.name, r.namespace).catch(() => {
    // ignore errors during uninstall ()
  })));

  await kubectlDelete(join(__dirname, "installer-local.yaml")); // needed for the console to start
  if (!options.skipCrd) {
    const crds = await getAllCRDs();
    await k8sDelete(crds.filter(crd => kymaCrds.includes(crd.metadata.name)));
  }
  if (options.deleteNamespaces) {
    debug("Deleting kyma namespaces");
    await kubectlDelete(join(__dirname, "system-namespaces.yaml"));
  }

  const usualLeftovers = [
    '/api/v1/namespaces/kyma-system/secrets',
    '/apis/oathkeeper.ory.sh/v1alpha1/namespaces/kyma-system/rules',
    '/apis/rafter.kyma-project.io/v1beta1/clusterassets',
    '/apis/rafter.kyma-project.io/v1beta1/clusterbuckets',
  ]
  await Promise.allSettled(usualLeftovers.map(path => deleteAllK8sResources(path)));
  const skipIstio = (skipComponents && skipComponents.includes("istio")) || (components && !components.includes("istio"));

  if (!skipIstio) {
    debug("Uninstalling istio");
    await uninstallIstio();
  }

}

/**
 * Install Kyma on kubernetes cluster with current kubeconfig
 * @param {Object} options List of installation options
 * @param {string} options.resourcesPath Path to the resources folder with Kyma charts 
 * @param {string} options.istioVersion Istio version, eg. 1.8.2
 * @param {boolean} options.isUpgrade Upgrade existing installation
 * @param {boolean} options.newEventing Use new eventing
 * @param {Array<string>} options.skipComponents List of components to not install
 * @param {Array<string>} options.components List of components to install
 * @param {boolean} options.isCompassEnabled
 * @param {boolean} options.useHelmTemplate Use "helm template | kubectl apply" instead of helm install/upgrade
 */
async function installKyma(options) {
  options = options || {};
  const installLocation = options.resourcesPath || join(__dirname, "..", "..", "..", "resources");
  const crdLocation = options.resourcesPath || join(__dirname, "..", "..", "..", "installation", "resources", "crds");
  const istioVersion = options.istioVersion || defaultIstioVersion;
  const isUpgrade = options.isUpgrade || false;
  const skipComponents = options.skipComponents || [];
  const components = options.components;
  console.time('Installation');
  console.log('Installing Kyma from folder', installLocation);
  await waitForNodesToBeReady();
  const crdsBefore = await getAllCRDs();
  const skipIstio = skipComponents.includes("istio") || (components && !components.includes("istio"));

  if (!skipIstio) {
    if (options.isUpgrade) {
      await upgradeIstio(istioVersion);
    } else {
      await updateCoreDNSConfigMap();
      await installIstio(istioVersion);
    }
  }
  console.timeLog('Installation', 'Istio installed');
  await removeKymaGatewayCertsYaml(installLocation);
  await kubectlApply(join(__dirname, "system-namespaces.yaml"));
  if (options.withCompass) {
    await kubectlApply(join(__dirname, "compass-namespace.yaml"));
  }
  await kubectlApplyDir(crdLocation, "kyma-system");
  let kymaCharts = await chartList(options);
  if (components) {
    kymaCharts = kymaCharts.filter(c => components.includes(c.release));
  }
  if (skipComponents) {
    kymaCharts = kymaCharts.filter(c => !skipComponents.includes(c.release))
  }
  const installFn = options.useHelmTemplate ? applyRelease : installRelease;
  await Promise.all(
    kymaCharts.map(({ release, namespace, values, customPath, profile }) => {
      const chartLocation = !!customPath
        ? customPath(installLocation)
        : join(installLocation, release);
      return pRetry(
        async () =>
          installFn(release, namespace, chartLocation, values, profile || "evaluation", isUpgrade),
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
  console.timeEnd('Installation');
}

module.exports = {
  installKyma,
  uninstallKyma,
  ...downloader
};
