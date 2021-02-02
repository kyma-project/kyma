const execa = require("execa");
const k8s = require("@kubernetes/client-node");
const { join } = require("path");
const { installIstio, upgradeIstio } = require("./istioctl");
const pRetry = require("p-retry");
const { debug } = require("../utils");
const notDeployed = "not-deployed";

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

// TODO: replace by kubectlApply from utils
async function kubectlApply(path, namespace) {
  const args = ["apply", "-f", path];
  if (!!namespace) {
    args.push("-n", namespace);
  }
  await execa("kubectl", args);
}

async function updateCoreDNSConfigMap() {
  const kc = new k8s.KubeConfig();
  kc.loadFromDefault();

  const cmName = "coredns";
  const cmNamespace = "kube-system";

  const k8sCoreV1Api = kc.makeApiClient(k8s.CoreV1Api);
  const { body } = await k8sCoreV1Api.readNamespacedConfigMap(
    "coredns",
    "kube-system"
  );

  const { stdout: registryIP } = await execa("docker", [
    "inspect",
    "-f",
    "{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}",
    "/registry.localhost",
  ]);

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

  if (!!values) {
    args.push("--set", values);
  }

  if (!!profile) {
    args.push("-f", join(chart, `profile-${profile}.yaml`));
  }

  await execa("helm", args);
}

async function installRelease(
  release,
  namespace = "kyma-system",
  chart,
  values,
  profile
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
      debug(`Upgrading release ${release}`);
      await helmInstallUpgrade(release, chart, namespace, values, profile);
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

const isGardener = process.env["GARDENER"] || "false";
const domain = process.env["KYMA_DOMAIN"] || "local.kyma.dev";

const overrides = `global.isLocalEnv=false,global.ingress.domainName=${domain},global.environment.gardener=${isGardener},global.domainName=${domain},global.tlsCrt=ZHVtbXkK`;

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
    profile: "evaluation",
  },
  {
    release: "serverless",
    namespace: "kyma-system",
    // https://github.com/kyma-project/test-infra/pull/2967
    values: `dockerRegistry.enableInternal=false,dockerRegistry.serverAddress=registry.localhost:5000,dockerRegistry.registryAddress=registry.localhost:5000,global.ingress.domainName=${domain},containers.manager.envs.functionBuildExecutorImage.value=eu.gcr.io/kyma-project/external/aerfio/kaniko-executor:v1.3.0`,
    profile: "evaluation",
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
    profile: "evaluation",
  },
  {
    release: "service-catalog",
    namespace: "kyma-system",
    values: `${overrides}`,
    profile: "evaluation",
  },
  {
    release: "service-catalog-addons",
    namespace: "kyma-system",
    values: `${overrides}`,
    profile: "evaluation",
  },
  {
    release: "helm-broker",
    namespace: "kyma-system",
    values: `${overrides}`,
    profile: "evaluation",
  },
  {
    release: "console",
    namespace: "kyma-system",
    values: `${overrides},pamela.enabled=false`,
    profile: "evaluation",
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
    values: `global.natsStreaming.resources.requests.memory=64M,global.natsStreaming.resources.requests.cpu=10m`,
  },
  {
    release: "event-sources",
    namespace: "kyma-system",
  },
  {
    release: "monitoring",
    namespace: "kyma-system",
    values: `${overrides}`,
    profile: "evaluation",
  },
  {
    release: "kiali",
    namespace: "kyma-system",
    values: `${overrides}`,
    profile: "evaluation",
  },
  {
    release: "tracing",
    namespace: "kyma-system",
    values: `${overrides}`,
    profile: "evaluation",
  },
  {
    release: "logging",
    namespace: "kyma-system",
    values: `${overrides}`,
    profile: "evaluation",
  },
  {
    release: "ingress-dns-cert",
    namespace: "istio-system",
    values: `global.ingress.domainName=${domain},global.environment.gardener=${isGardener}`,
    customPath: () => join(__dirname, "charts", "ingress-dns-cert"),
  },
];

async function installKyma(
  installLocation = join(__dirname, "..", "..", "..", "resources"),
  istioVersion,
  ignoredComponents = "monitoring,tracing,kiali,logging,console,cluster-users,dex",
  isUpgrade = false
) {
  await waitForNodesToBeReady();

  if (isUpgrade) {
    await upgradeIstio(istioVersion);
  } else {
    await updateCoreDNSConfigMap();
    await installIstio(istioVersion);
  }

  await removeKymaGatewayCertsYaml(installLocation);
  await kubectlApply(join(__dirname, "installer-local.yaml")); // needed for the console to start
  await kubectlApply(join(__dirname, "system-namespaces.yaml"));
  await kubectlApply(
    join(installLocation, "cluster-essentials/files"),
    "kyma-system"
  );

  return await Promise.all(
    kymaCharts
      .filter((arg) => !ignoredComponents.includes(arg.release))
      .map(({ release, namespace, values, customPath, profile }) => {
        const chartLocation = !!customPath
          ? customPath(installLocation)
          : join(installLocation, release);
        return pRetry(
          async () =>
            installRelease(release, namespace, chartLocation, values, profile),
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
}

module.exports = {
  installKyma,
};
