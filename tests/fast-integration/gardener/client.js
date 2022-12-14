const k8s = require('@kubernetes/client-node');
const {fromBase64, getEnvOrThrow, debug, error} = require('../utils');

const GARDENER_PROJECT = process.env.KCP_GARDENER_NAMESPACE || 'garden-kyma-dev';
const COMPASS_ID_ANNOTATION_KEY = 'compass.provisioner.kyma-project.io/runtime-id';

const {
  waitForK8sObject,
} = require('../utils');

class GardenerConfig {
  static fromEnv() {
    return new GardenerConfig({
      kubeconfigPath: getEnvOrThrow('GARDENER_KUBECONFIG'),
      // Exception is not necessary below - shootTemplate is not used in all tests
      shootTemplate: process.env['GARDENER_SHOOT_TEMPLATE'],
    });
  }

  static fromString(kubeconfig) {
    return new GardenerConfig({
      kubeconfig: kubeconfig,
    });
  }

  constructor(opts) {
    opts = opts || {};
    this.kubeconfigPath = opts.kubeconfigPath;
    this.kubeconfig = opts.kubeconfig;
    this.shootTemplate = opts.shootTemplate;
  }
}

class GardenerClient {
  constructor(config) {
    config = config || {};
    const kc = new k8s.KubeConfig();
    if (config.kubeconfigPath) {
      kc.loadFromFile(config.kubeconfigPath);
    } else if (config.kubeconfig) {
      kc.loadFromString(config.kubeconfig);
    } else {
      throw new Error('Unable to create GardenerClient - no kubeconfig');
    }

    this.coreV1API = kc.makeApiClient(k8s.CoreV1Api);
    this.dynamicAPI = kc.makeApiClient(k8s.KubernetesObjectApi);
    this.watch = new k8s.Watch(kc);
    this.shootTemplate = config.shootTemplate;
  }

  async createShoot(shootName) {
    debug(`Creating a K8S cluster in gardener namespace`);
    if (!this.shootTemplate) {
      error(`No shoot Template`);
      return new Error(`no shoot template defined in the Gardener client`);
    }

    const data = fromBase64(this.shootTemplate);

    let replaced = data.replace(/<SHOOT>/g, shootName);
    replaced = replaced.replace(/<NAMESPACE>/g, GARDENER_PROJECT);

    const shootTemplate = k8s.loadYaml(replaced);

    await this.dynamicAPI.create(shootTemplate)
        .catch((err) => {
          const message = JSON.stringify(err.body.message);
          error(`Got the error with response ${message}`);
        });

    await this.waitForShoot(shootName);

    return this.getShoot(shootName);
  }

  async deleteShoot(name) {
    await this.dynamicAPI.delete({
      apiVersion: 'core.gardener.cloud/v1beta1',
      kind: 'Shoot',
      metadata: {
        name: name,
        namespace: GARDENER_PROJECT,
      },
    })
        .catch((err) => {
          const message = JSON.stringify(err.body.message);
          error(`Got the error with response ${message}`);
        });
  }

  async waitForShoot(shootName) {
    return waitForK8sObject(
        `/apis/core.gardener.cloud/v1beta1/namespaces/${GARDENER_PROJECT}/shoots`, {}, (_type, _apiObj, watchObj) => {
          if (watchObj.object.metadata.name != shootName) {
            return false;
          }

          debug(`Waiting for ${watchObj.object.metadata.name} shoot`);
          const lastOperation = watchObj.object.status.lastOperation;

          return lastOperation.type == 'Create' && lastOperation.state == 'Succeeded';
        }, 1200 * 1000, 'Waiting for shoot to be ready timeout', this.watch);
  }

  async getShoot(shootName) {
    debug(`Fetching shoot: ${shootName} from gardener namespace: ${GARDENER_PROJECT}`);

    const secretResp = await this.coreV1API.readNamespacedSecret(
        `${shootName}.kubeconfig`,
        GARDENER_PROJECT,
    );

    const shootResp = await this.dynamicAPI.read({
      apiVersion: 'core.gardener.cloud/v1beta1',
      kind: 'Shoot',
      metadata: {
        name: shootName,
        namespace: GARDENER_PROJECT,
      },
    });

    return {
      name: shootName,
      compassID: shootResp.body.metadata.annotations[COMPASS_ID_ANNOTATION_KEY],
      kubeconfig: fromBase64(secretResp.body.data.kubeconfig),
      oidcConfig: shootResp.body.spec.kubernetes.kubeAPIServer.oidcConfig,
      shootDomain: shootResp.body.spec.dns.domain,
    };
  }
}

module.exports = {
  GardenerConfig,
  GardenerClient,
};
