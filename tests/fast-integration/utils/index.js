const stream = require('stream');
const k8s = require('@kubernetes/client-node');
const fs = require('fs');
const {join} = require('path');
const {expect} = require('chai');
const execa = require('execa');

const kc = new k8s.KubeConfig();
let k8sDynamicApi;
let k8sAppsApi;
let k8sCoreV1Api;
let k8sRbacAuthorizationV1Api;
let k8sLog;
let k8sServerUrl;

let watch;

const eventingBackendName = 'eventing-backend';

function initializeK8sClient(opts) {
  opts = opts || {};
  try {
    console.log('Trying to initialize a K8S client');
    if (opts.kubeconfigPath) {
      console.log('Path initialization');
      kc.loadFromFile(opts.kubeconfigPath);
    } else if (opts.kubeconfig) {
      console.log('Kubeconfig initialization');
      kc.loadFromString(opts.kubeconfig);
    } else {
      console.log('Default initialization');
      kc.loadFromDefault();
    }

    console.log('Clients creation');
    k8sDynamicApi = kc.makeApiClient(k8s.KubernetesObjectApi);
    console.log('Making Api client - Apps');
    k8sAppsApi = kc.makeApiClient(k8s.AppsV1Api);
    console.log('Making Api client - Core');
    k8sCoreV1Api = kc.makeApiClient(k8s.CoreV1Api);
    console.log('Making Api client - Auth');
    k8sRbacAuthorizationV1Api = kc.makeApiClient(k8s.RbacAuthorizationV1Api);
    console.log('Making Api client - Logs');
    k8sLog = new k8s.Log(kc);
    console.log('Making Api client - Watch');
    watch = new k8s.Watch(kc);
    k8sServerUrl = kc.getCurrentCluster() ? kc.getCurrentCluster().server : null;
  } catch (err) {
    console.log(err.message);
  }
}
initializeK8sClient();

/**
 * Gets the shoot name from k8s server url
 *
 * @throws
 * @return {String}
 */
function getShootNameFromK8sServerUrl() {
  if (!k8sServerUrl || k8sServerUrl === '' || k8sServerUrl.split('.').length < 1) {
    throw new Error(`failed to get shootName from K8s server Url: ${k8sServerUrl}`);
  }
  return k8sServerUrl.split('.')[1];
}


/**
 * Retries a promise {retriesLeft} times every {interval} miliseconds
 *
 * @async
 * @param {function() : Promise} fn - async function that returns a promise
 * @param {number=} retriesLeft
 * @param {number=} interval
 * @throws
 * @return {Promise}
 */
async function retryPromise(fn, retriesLeft = 10, interval = 30) {
  if (retriesLeft < 1) {
    throw new Error('retriesLeft argument should be greater then 0');
  }

  return new Promise((resolve, reject) => {
    return fn()
        .then(resolve)
        .catch((error) => {
          if (retriesLeft === 1) {
          // reject('maximum retries exceeded');
            reject(error);
            return;
          }
          setTimeout(() => {
          // Passing on "reject" is the important part
            retryPromise(fn, retriesLeft - 1, interval).then(resolve, reject);
          }, interval);
        });
  });
}

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}


function escapeRegExp(str) {
  return str.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'); // $& means the whole matched string
}
function replaceAllInString(str, match, replacement) {
  return str.replace(new RegExp(escapeRegExp(match), 'g'), () => replacement);
}

function convertAxiosError(axiosError, message) {
  if (!axiosError.response) {
    return new Error(`${message}: ${axiosError.toString()}`);
  }
  if (
    axiosError.response &&
    axiosError.response.status &&
    axiosError.response.statusText
  ) {
    message += `\n${axiosError.response.status}: ${axiosError.response.statusText}`;
  }
  if (axiosError.response && axiosError.response.data) {
    message += ': ' + JSON.stringify(axiosError.response.data);
  }
  return new Error(message);
}

// "polyfill" for Promise.allSettled
async function promiseAllSettled(promises) {
  return Promise.all(
      promises.map((promise) =>
        promise
            .then((value) => ({
              status: 'fulfilled',
              value,
            }))
            .catch((reason) => ({
              status: 'rejected',
              reason,
            })),
      ),
  );
}

/**
 *
 * @param {string} dir path to the directory with yaml files
 * @param {string} namespace to apply yamls to
 */
async function kubectlApplyDir(dir, namespace) {
  const files = fs.readdirSync(dir);
  for (const file of files) {
    if (file.endsWith('.yaml') || file.endsWith('.yml')) {
      await kubectlApply(join(dir, file), namespace).catch(console.error);
    } else {
      await kubectlApplyDir(join(dir, file), namespace);
    }
  }
}

async function kubectlApply(file, namespace) {
  const yaml = fs.readFileSync(file);
  const resources = k8s.loadAllYaml(yaml);
  await k8sApply(resources, namespace);
}

async function kubectlDeleteDir(dir, namespace) {
  const files = fs.readdirSync(dir);
  for (const file of files) {
    if (file.endsWith('.yaml') || file.endsWith('.yml')) {
      await kubectlDelete(join(dir, file), namespace).catch(console.error);
    }
  }
}

function kubectlDelete(file, namespace) {
  debug(`Deleting ${file}...`);
  const yaml = fs.readFileSync(file);
  const listOfSpecs = k8s.loadAllYaml(yaml);
  return k8sDelete(listOfSpecs, namespace);
}

async function k8sDelete(listOfSpecs, namespace) {
  for (const res of listOfSpecs) {
    if (namespace) {
      res.metadata.namespace = namespace;
    }
    debug(`Delete ${res.metadata.name}`);
    try {
      if (res.kind) {
        await k8sDynamicApi.delete(res);
      } else if (res.metadata.selfLink) {
        await k8sDynamicApi.requestPromise({
          url: k8sDynamicApi.basePath + res.metadata.selfLink,
          method: 'DELETE',
        });
      } else {
        throw Error(
            'Object kind or metadata.selfLink is required to delete the resource',
        );
      }
      if (res.kind === 'CustomResourceDefinition') {
        const version = res.spec.version || res.spec.versions[0].name;
        const path = `/apis/${res.spec.group}/${version}/${res.spec.names.plural}`;
        await deleteAllK8sResources(path);
      }
    } catch (err) {
      ignore404(err);
    }
  }
}

async function getAllCRDs() {
  const path = '/apis/apiextensions.k8s.io/v1/customresourcedefinitions';
  const response = await k8sDynamicApi.requestPromise({
    url: k8sDynamicApi.basePath + path,
  });
  const stat = {};
  const body = JSON.parse(response.body);
  body.items.forEach(
      (crd) =>
        (stat[crd.spec.group] = stat[crd.spec.group] ?
        stat[crd.spec.group] + 1 :
        1),
  );
  return body.items;
}

async function getClusteraddonsconfigurations() {
  const path =
    '/apis/addons.kyma-project.io/v1alpha1/clusteraddonsconfigurations';
  const response = await k8sDynamicApi.requestPromise({
    url: k8sDynamicApi.basePath + path,
  });
  const body = JSON.parse(response.body);
  return body.items;
}

async function getSecrets(namespace) {
  const path = `/api/v1/namespaces/${namespace}/secrets`;
  const response = await k8sDynamicApi.requestPromise({
    url: k8sDynamicApi.basePath + path,
  });
  const body = JSON.parse(response.body);
  return body.items;
}

async function getSecret(name, namespace) {
  const path = `/api/v1/namespaces/${namespace}/secrets/${name}`;
  const response = await k8sDynamicApi.requestPromise({
    url: k8sDynamicApi.basePath + path,
  });
  return JSON.parse(response.body);
}

async function getFunction(name, namespace) {
  const path = `/apis/serverless.kyma-project.io/v1alpha2/namespaces/${namespace}/functions/${name}`;
  const response = await k8sDynamicApi.requestPromise({
    url: k8sDynamicApi.basePath + path,
  });
  return JSON.parse(response.body);
}

async function getConfigMap(name, namespace='default') {
  const path = `/api/v1/namespaces/${namespace}/configmaps/${name}`;
  const response = await k8sDynamicApi.requestPromise({
    url: k8sDynamicApi.basePath + path,
  });
  return JSON.parse(response.body);
}

async function k8sApply(resources, namespace, patch = true) {
  const options = {
    headers: {'Content-type': 'application/merge-patch+json'},
  };
  for (const resource of resources) {
    if (!resource || !resource.kind || !resource.metadata.name) {
      debug('Skipping invalid resource:', resource);
      continue;
    }
    if (!resource.metadata.namespace) {
      resource.metadata.namespace = namespace;
    }
    if (resource.kind == 'Namespace') {
      resource.metadata.labels = {
        'istio-injection': 'enabled',
      };
    }
    try {
      await k8sDynamicApi.patch(
          resource,
          undefined,
          undefined,
          undefined,
          undefined,
          options,
      );
      debug(resource.kind, resource.metadata.name, 'reconfigured');
    } catch (e) {
      {
        if (e.body && e.body.reason === 'NotFound') {
          try {
            await k8sDynamicApi.create(resource);
            debug(resource.kind, resource.metadata.name, 'created');
          } catch (createError) {
            debug(resource.kind, resource.metadata.name, 'failed to create');
            debug(JSON.stringify(createError, null, 4));
            throw createError;
          }
        } else {
          throw e;
        }
      }
    }
  }
}

// Allows to pass watch with different than global K8S context.
function waitForK8sObject(path, query, checkFn, timeout, timeoutMsg, watcher = watch) {
  debug('waiting for', path);
  let res;
  let timer;
  return new Promise((resolve, reject) => {
    watcher.watch(
        path,
        query,
        (type, apiObj, watchObj) => {
          if (checkFn(type, apiObj, watchObj)) {
            if (res) {
              res.abort();
            }
            clearTimeout(timer);
            debug('finished waiting for ', path);
            resolve(watchObj.object);
          }
        },
        () => {
        },
    )
        .then((r) => {
          res = r;
          timer = setTimeout(() => {
            res.abort();
            reject(new Error(timeoutMsg));
          }, timeout);
        });
  });
}

function waitForNamespace(name, timeout = 30_000) {
  return waitForK8sObject(
      `/api/v1/namespaces/${name}`,
      {},
      (_type, _apiObj, watchObj) => {
        return (
          watchObj.metadata.name === name &&
            watchObj.status.phase === 'Active'
        );
      },
      timeout,
      `Waiting for ${name} namespace timeout 3000 ms)`,
  );
}

function waitForClusterAddonsConfiguration(name, timeout = 90_000) {
  return waitForK8sObject(
      '/apis/addons.kyma-project.io/v1alpha1/clusteraddonsconfigurations',
      {},
      (_type, _apiObj, watchObj) => {
        return watchObj.object.metadata.name === name;
      },
      timeout,
      `Waiting for ${name} ClusterAddonsConfiguration timeout (${timeout} ms)`,
  );
}

function waitForApplicationCr(appName, timeout = 300_000) {
  return waitForK8sObject(
      '/apis/applicationconnector.kyma-project.io/v1alpha1/applications',
      {},
      (_type, _apiObj, watchObj) => {
        return (
          watchObj.object.metadata.name == appName
        );
      },
      timeout,
      `Waiting for application ${appName} timeout (${timeout} ms)`,
  );
}

function waitForEndpoint(name, namespace = 'default', timeout = 300_000) {
  return waitForK8sObject(
      `/api/v1/namespaces/${namespace}/endpoints`,
      {},
      (_type, _apiObj, watchObj) => {
        return (
          watchObj.object.metadata.name === name &&
              watchObj.object.subsets
        );
      },
      timeout,
      `Waiting for endpoint ${name} timeout (${timeout} ms)`,
  );
}

function waitForFunction(name, namespace = 'default', timeout = 90_000) {
  return waitForK8sObject(
      `/apis/serverless.kyma-project.io/v1alpha2/namespaces/${namespace}/functions`,
      {},
      (_type, _apiObj, watchObj) => {
        return (
          watchObj.object.metadata.name === name &&
        watchObj.object.status.conditions &&
        watchObj.object.status.conditions.some(
            (c) => c.type === 'Running' && c.status === 'True',
        )
        );
      },
      timeout,
      `Waiting for ${name} function timeout (${timeout} ms)`,
  );
}

async function getSubscription(name, namespace = 'default', crdVersion='v1alpha1') {
  try {
    const path = `/apis/eventing.kyma-project.io/${crdVersion}/namespaces/${namespace}/subscriptions/${name}`;
    const response = await k8sDynamicApi.requestPromise({
      url: k8sDynamicApi.basePath + path,
      qs: {limit: 500},
    });
    const body = JSON.parse(response.body);
    return body;
  } catch (e) {
    if (e.statusCode === 404 || e.statusCode === 405) {
      // do nothing
    } else {
      error(e);
      throw e;
    }
  }
}

async function getAllSubscriptions(namespace = 'default', crdVersion='v1alpha1') {
  try {
    const path = `/apis/eventing.kyma-project.io/${crdVersion}/namespaces/${namespace}/subscriptions`;
    const response = await k8sDynamicApi.requestPromise({
      url: k8sDynamicApi.basePath + path,
      qs: {limit: 500},
    });
    const body = JSON.parse(response.body);

    return Promise.all(
        body.items.map((sub) => {
          return {
            apiVersion: sub['apiVersion'],
            spec: sub['spec'],
            status: sub['status'],

          };
        }),
    ).then((results) => {
      return results.flat();
    });
  } catch (e) {
    if (e.statusCode === 404 || e.statusCode === 405) {
      // do nothing
    } else {
      error(e);
      throw e;
    }
  }
}

// gets the active eventing backend
async function getEventingBackend(namespace = 'kyma-system') {
  const path = '/apis/eventing.kyma-project.io/v1alpha1/eventingbackends/';
  const response = await k8sDynamicApi.requestPromise({
    url: k8sDynamicApi.basePath + path,
  });
  const body = JSON.parse(response.body);
  for (let i = 0; i < body.items.length; i++) {
    const item = body.items[i];
    if (item?.metadata?.name === eventingBackendName) {
      return item?.status?.backendType;
    }
  }
  return '';
}

function waitForSubscription(name, namespace = 'default', crdVersion='v1alpha1', timeout = 180_000) {
  return waitForK8sObject(
      `/apis/eventing.kyma-project.io/${crdVersion}/namespaces/${namespace}/subscriptions`,
      {},
      (_type, _apiObj, watchObj) => {
        return (
          watchObj.object.metadata.name === name &&
        watchObj.object.status.conditions &&
        watchObj.object.status.conditions.some(
            (c) => c.type === 'Subscription active' && c.status === 'True',
        )
        );
      },
      timeout,
      `Waiting for ${name} subscription timeout (${timeout} ms)`,
  );
}

function waitForReplicaSet(name, namespace = 'default', timeout = 90_000) {
  return waitForK8sObject(
      `/apis/apps/v1/namespaces/${namespace}/replicasets`,
      {},
      (_type, _apiObj, watchObj) => {
        return (
          watchObj.object.metadata.name === name &&
        watchObj.object.status &&
        watchObj.object.status.readyReplicas > 0 &&
        watchObj.object.status.availableReplicas > 0
        );
      },
      timeout,
      `Waiting for replica set ${name} timeout (${timeout} ms)`,
  );
}

function waitForDaemonSet(name, namespace = 'default', timeout = 90_000) {
  return waitForK8sObject(
      `/apis/apps/v1/watch/namespaces/${namespace}/daemonsets/${name}`,
      {},
      (_type, watchObj, _) => {
        return (
          watchObj.status.numberReady === watchObj.status.desiredNumberScheduled
        );
      },
      timeout,
      `Waiting for daemonset ${name} timeout (${timeout} ms)`,
  );
}

function waitForDeployment(name, namespace = 'default', timeout = 90_000) {
  return waitForK8sObject(
      `/apis/apps/v1/namespaces/${namespace}/deployments`,
      {},
      (_type, _apiObj, watchObj) => {
        return (
          watchObj.object.metadata.name === name &&
        watchObj.object.status.conditions &&
        watchObj.object.status.conditions.some(
            (c) => c.type === 'Available' && c.status === 'True',
        )
        );
      },
      timeout,
      `Waiting for deployment ${name} timeout (${timeout} ms)`,
  );
}

function waitForService(name, namespace = 'default', timeout = 90_000) {
  return waitForK8sObject(
      `/api/v1/namespaces/${namespace}/services`,
      {},
      (_type, _apiObj, watchObj) => {
        return (
          watchObj.object.metadata.name === name &&
          watchObj.object.spec.clusterIP
        );
      },
      timeout,
      `Waiting for service ${name} timeout (${timeout} ms)`,
  );
}

function waitForStatefulSet(name, namespace = 'default', timeout = 90_000) {
  return waitForK8sObject(
      `/apis/apps/v1/namespaces/${namespace}/statefulsets`,
      {},
      (_type, _apiObj, watchObj) => {
        return (
          watchObj.object.metadata.name === name &&
        watchObj.object.status.readyReplicas > 0
        );
      },
      timeout,
      `Waiting for StatefulSet ${name} timeout (${timeout} ms)`,
  );
}

function waitForJob(name, namespace = 'default', timeout = 900_000, success = 1) {
  return waitForK8sObject(
      `/apis/batch/v1/namespaces/${namespace}/jobs`,
      {},
      (_type, _apiObj, watchObj) => {
        return (
          watchObj.object.metadata.name === name &&
        watchObj.object.status.succeeded >= success
        );
      },
      timeout,
      `Waiting for Job ${name} to succeed ${success} timeout (${timeout} ms)`,
  );
}

async function kubectlExecInPod(pod, container, cmd, namespace = 'default', timeoutInSeconds = 60) {
  const execCmd = ['exec', pod, '-c', container, '-n', namespace, '--', ...cmd];

  for (let i = 0; i < timeoutInSeconds/10; i++) {
    try {
      await execa(`kubectl`, execCmd);
      console.log(`kubectl command ${execCmd.join(' ')} executed`);
      return;
    } catch (error) {
      if (i === timeoutInSeconds/10-1) {
        if (error.stdout === undefined) {
          throw error;
        }
        throw new Error(`failed to execute kubectl ${execCmd.join(' ')}:\n${error.stdout},\n${error.stderr}`);
      }
      console.log(`Retry attempt: ${i} Failed to execute kubectl ${execCmd.join(' ')}:\n${error.stdout},
        ${error.stderr}`);
    }
    await sleep(10000);
  }
}

async function listPods(selector, namespace = 'default') {
  return await k8sCoreV1Api.listNamespacedPod(namespace, undefined, undefined, undefined, undefined, selector);
}

async function printContainerLogs(selector, container, namespace = 'default', timeout = 90000) {
  const res = await k8sCoreV1Api.listNamespacedPod(namespace, undefined, undefined, undefined, undefined, selector);
  res.body.items.sort((a, b) => a.metadata.creationTimestamp - b.metadata.creationTimestamp);
  for (const p of res.body.items) {
    process.stdout.write(`Getting logs for pod ${p.metadata.name}/${container}\n`);
    const logStream = new stream.PassThrough();
    logStream.on('data', (chunk) => {
      // use write rather than console.log to prevent double line feed
      process.stdout.write(chunk);
    });
    const end = new Promise(function(resolve, reject) {
      logStream.on('end', () => {
        process.stdout.write('\n'); resolve();
      });
      logStream.on('error', reject);
    });
    k8sLog.log(namespace, p.metadata.name, container, logStream);
    await end;
  }
  process.stdout.write('Done getting logs\n');
}

function waitForVirtualService(namespace, apiRuleName, timeout = 30_000) {
  const path = `/apis/networking.istio.io/v1beta1/namespaces/${namespace}/virtualservices`;
  const query = {
    labelSelector: `apirule.gateway.kyma-project.io/v1beta1=${apiRuleName}.${namespace}`,
  };
  return waitForK8sObject(
      path,
      query,
      (_type, _apiObj, watchObj) => {
        return (
          watchObj.object.spec.hosts && watchObj.object.spec.hosts.length === 1
        );
      },
      timeout,
      `Wait for VirtualService ${apiRuleName} timeout (${timeout} ms)`,
  );
}

async function getVirtualService(namespace, name) {
  try {
    const path = `/apis/networking.istio.io/v1beta1/namespaces/${namespace}/virtualservices/${name}`;
    const response = await k8sDynamicApi.requestPromise({
      url: k8sDynamicApi.basePath + path,
    });
    return JSON.parse(response.body);
  } catch (err) {
    return JSON.parse(err.response.body);
  }
}

async function getGateway(namespace, name) {
  try {
    const path = `/apis/networking.istio.io/v1beta1/namespaces/${namespace}/gateways/${name}`;
    const response = await k8sDynamicApi.requestPromise({
      url: k8sDynamicApi.basePath + path,
    });
    return JSON.parse(response.body);
  } catch (err) {
    return JSON.parse(err.response.body);
  }
}

async function getPersistentVolumeClaim(namespace, name) {
  const path = `/api/v1/namespaces/${namespace}/persistentvolumeclaims/${name}`;
  const response = await k8sDynamicApi.requestPromise({
    url: k8sDynamicApi.basePath + path,
  });
  return JSON.parse(response.body);
}

function waitForTokenRequest(name, namespace, timeout = 5000) {
  const path = `/apis/applicationconnector.kyma-project.io/v1alpha1/namespaces/${namespace}/tokenrequests`;
  return waitForK8sObject(
      path,
      {},
      (_type, _apiObj, watchObj) => {
        return (
          watchObj.object.metadata.name === name &&
        watchObj.object.status &&
        watchObj.object.status.state === 'OK' &&
        watchObj.object.status.url
        );
      },
      timeout,
      'Wait for TokenRequest timeout',
  );
}

function waitForCompassConnection(name, timeout = 90000) {
  const path = '/apis/compass.kyma-project.io/v1alpha1/compassconnections';
  return waitForK8sObject(
      path,
      {},
      (_type, _apiObj, watchObj) => {
        return (
          watchObj.object.metadata.name === name &&
        watchObj.object.status.connectionState &&
        ['Connected', 'Synchronized'].indexOf(
            watchObj.object.status.connectionState,
        ) !== -1
        );
      },
      timeout,
      `Wait for Compass connection ${name} timeout (${timeout} ms)`,
  );
}

function waitForPodWithLabel(
    labelKey,
    labelValue,
    namespace = 'default',
    timeout = 90000,
) {
  const query = {
    labelSelector: `${labelKey}=${labelValue}`,
  };
  return waitForK8sObject(
      `/api/v1/namespaces/${namespace}/pods`,
      query,
      (_type, _apiObj, watchObj) => {
        return (
          watchObj.object.status.phase === 'Running' &&
        watchObj.object.status.containerStatuses.every((cs) => cs.ready)
        );
      },
      timeout,
      `Waiting for pod with label ${labelKey}=${labelValue} timeout (${timeout} ms)`,
  );
}

function waitForPodWithLabelAndCondition(
    labelKey,
    labelValue,
    namespace = 'default',
    condition = 'Ready',
    conditionStatus = 'True',
    timeout = 90_000,
) {
  const query = {
    labelSelector: `${labelKey}=${labelValue}`,
  };
  return waitForK8sObject(
      `/api/v1/namespaces/${namespace}/pods`,
      query,
      (_type, _apiObj, watchObj) => {
        debug(`Waiting for pod "${namespace}/${watchObj.object.metadata.name}" 
          to have condition "${condition}: ${conditionStatus}" for ${timeout} ms`);
        return (
          watchObj.object.status.conditions &&
            watchObj.object.status.conditions.some(
                (c) => c.type === condition && c.status === conditionStatus,
            )
        );
      },
      timeout,
      `Waiting for pod with label ${labelKey}=${labelValue} 
      and condition ${condition}=${conditionStatus} timeout (${timeout} ms)`,
  );
}

function waitForPodStatusWithLabel(
    labelKey,
    labelValue,
    namespace = 'default',
    status = 'Running',
    timeout = 90_000,
) {
  const query = {
    labelSelector: `${labelKey}=${labelValue}`,
  };
  return waitForK8sObject(
      `/api/v1/namespaces/${namespace}/pods`,
      query,
      (_type, _apiObj, watchObj) => {
        debug(`Waiting for pod "${namespace}/${watchObj.object.metadata.name}" status "${status}"`);
        return watchObj.object.status.phase === status;
      },
      timeout,
      `Waiting for pod status ${status} with label ${labelKey}=${labelValue} timeout (${timeout} ms)`,
  );
}

function waitForConfigMap(
    cmName,
    namespace = 'default',
    timeout = 90_000,
) {
  return waitForK8sObject(
      `/api/v1/namespaces/${namespace}/configmaps`,
      {},
      (_type, _apiObj, watchObj) => {
        return watchObj.object.metadata.name.includes(
            cmName,
        );
      },
      timeout,
      `Waiting for ${cmName} ConfigMap timeout (${timeout} ms)`,
  );
}

function waitForSecret(
    secretName,
    namespace = 'default',
    timeout = 90_000,
) {
  return waitForK8sObject(
      `/api/v1/namespaces/${namespace}/secrets`,
      {},
      (_type, _apiObj, watchObj) => {
        return watchObj.object.metadata.name.includes(
            secretName,
        );
      },
      timeout,
      `Waiting for ${secretName} Secret timeout (${timeout} ms)`,
  );
}

async function deleteNamespaces(namespaces, wait = true) {
  const result = await k8sCoreV1Api.listNamespace();
  const allNamespaces = result.body.items.map((i) => i.metadata.name);
  namespaces = namespaces.filter((n) => allNamespaces.includes(n));
  if (namespaces.length === 0) {
    return;
  }
  const waitForNamespacesResult = waitForK8sObject(
      '/api/v1/namespaces',
      {},
      (type, _, watchObj) => {
        if (type === 'DELETED') {
          namespaces = namespaces.filter(
              (n) => n !== watchObj.object.metadata.name,
          );
        }
        return namespaces.length === 0 || !wait;
      },
      10 * 60 * 1000,
      'Timeout for deleting namespaces: ' + namespaces,
  );

  for (const name of namespaces) {
    k8sDynamicApi
        .delete({
          apiVersion: 'v1',
          kind: 'Namespace',
          metadata: {name},
        })
        .catch(ignoreNotFound);
  }
  return waitForNamespacesResult;
}

async function listResources(path) {
  try {
    const listResponse = await k8sDynamicApi.requestPromise({
      url: k8sDynamicApi.basePath + path,
    });
    const listObj = JSON.parse(listResponse.body);
    if (listObj.items) {
      return listObj.items;
    }
  } catch (e) {
    if (e.statusCode !== 404 && e.statusCode !== 405) {
      console.error('Error:', e);
      throw e;
    }
  }
  return [];
}

async function listResourceNames(path) {
  const resources = await listResources(path);
  return resources.map((o) => o.metadata.name);
}

async function resourceTypes(group, version) {
  const path = group ? `/apis/${group}/${version}` : `/api/${version}`;
  try {
    const response = await k8sDynamicApi.requestPromise({
      url: k8sDynamicApi.basePath + path,
      qs: {limit: 500},
    });
    const body = JSON.parse(response.body);

    return body.resources.map((res) => {
      return {group, version, path, ...res};
    });
  } catch (e) {
    if (e.statusCode !== 404 && e.statusCode !== 405) {
      console.log('Error:', e);
      throw e;
    }
    return [];
  }
}

async function getAllResourceTypes() {
  try {
    const path = '/apis/apiregistration.k8s.io/v1/apiservices';
    const response = await k8sDynamicApi.requestPromise({
      url: k8sDynamicApi.basePath + path,
      qs: {limit: 500},
    });
    const body = JSON.parse(response.body);
    return Promise.all(
        body.items.map((api) => {
          return resourceTypes(api.spec.group, api.spec.version);
        }),
    ).then((results) => {
      return results.flat();
    });
  } catch (e) {
    if (e.statusCode === 404 || e.statusCode === 405) {
      // do nothing
    } else {
      console.log('Error:', e);
      throw e;
    }
  }
}

async function getSecretData(name, namespace) {
  try {
    const secret = await getSecret(name, namespace);
    const encodedData = secret.data;
    return Object.fromEntries(
        Object.entries(encodedData).map(([key, value]) => {
          const buff = Buffer.from(value, 'base64');
          const decoded = buff.toString('ascii');
          return [key, decoded];
        }),
    );
  } catch (e) {
    console.log('Error:', e);
    throw e;
  }
}

function ignore404(e) {
  if (
    (e.statusCode && e.statusCode === 404) ||
      (e.response && e.response.statusCode && e.response.statusCode === 404)
  ) {
    debug('Warning: Ignoring NotFound error');
    return;
  }

  throw e;
}

function ignoreNotFound(e) {
  if (e.body && e.body.reason === 'NotFound') {
    return;
  }

  console.log(e.body);
  throw e;
}

// NOTE: this works only for those where resource == lowercase plural kind
async function deleteK8sObjects(objects) {
  console.log(`deleting ${objects.length} objects`);
  for (const o of objects) {
    const path = `${o.apiVersion}/namespaces/${o.metadata.namespace}/${o.kind.toLowerCase()}s/${o.metadata.name}`;
    await k8sDynamicApi.requestPromise({
      url: `${k8sDynamicApi.basePath}/apis/${path}`,
      method: 'DELETE',
    });
  }
}

// NOTE: this no longer works, it relies on kube-api sending `selfLink` but the field has been deprecated
async function deleteAllK8sResources(
    path,
    query = {},
    retries = 2,
    interval = 1000,
    keepFinalizer = false,
) {
  try {
    let i = 0;
    while (i < retries) {
      if (i++) {
        await sleep(interval);
      }
      const response = await k8sDynamicApi.requestPromise({
        url: k8sDynamicApi.basePath + path,
        qs: query,
      });
      const body = JSON.parse(response.body);
      if (body.items && body.items.length) {
        for (const o of body.items) {
          await deleteK8sResource(o, path, keepFinalizer);
        }
      } else if (!body.items) {
        await deleteK8sResource(body, path, keepFinalizer);
      }
    }
  } catch (e) {
    debug('Error during delete ', path, String(e).substring(0, 1000));
    debug(e);
  }
}

async function deleteK8sPod(o) {
  return await k8sCoreV1Api.deleteNamespacedPod(o.metadata.name, o.metadata.namespace);
}

async function deleteK8sResource(o, path, keepFinalizer = false) {
  if (o.metadata.finalizers && o.metadata.finalizers.length && !keepFinalizer) {
    const options = {
      headers: {'Content-type': 'application/merge-patch+json'},
    };

    const obj = {
      kind: o.kind || 'Secret', // Secret list doesn't return kind and apiVersion
      apiVersion: o.apiVersion || 'v1',
      metadata: {
        name: o.metadata.name,
        namespace: o.metadata.namespace,
        finalizers: [],
      },
    };

    debug('Removing finalizers from', obj);
    try {
      await k8sDynamicApi.patch(obj, undefined, undefined, undefined, undefined, options);
    } catch (err) {
      ignore404(err);
    }
  }

  try {
    let objectUrl = `${k8sDynamicApi.basePath + path}/${o.metadata.name}`;
    if (o.metadata.selfLink) {
      debug('using selfLink for deleting object');
      objectUrl = k8sDynamicApi.basePath + o.metadata.selfLink;
    }

    debug('Deleting resource: ', objectUrl);
    await k8sDynamicApi.requestPromise({
      url: objectUrl,
      method: 'DELETE',
    });
  } catch (err) {
    ignore404(err);
  }

  debug(
      'Deleted resource:',
      o.metadata.name,
      'namespace:',
      o.metadata.namespace,
  );
}

async function getContainerRestartsForAllNamespaces() {
  const {body} = await k8sCoreV1Api.listPodForAllNamespaces();
  const pods = body.items;
  return pods
      .filter((pd) => !!pd.status && !!pd.status.containerStatuses)
      .map((pod) => ({
        name: pod.metadata.name,
        containerStatuses: pod.status.containerStatuses.map((elem) => ({
          name: elem.name,
          image: elem.image,
          restartCount: elem.restartCount,
        })),
      }));
}

async function getKymaAdminBindings() {
  const {body} = await k8sRbacAuthorizationV1Api.listClusterRoleBinding();
  const adminRoleBindings = body.items;
  return adminRoleBindings
      .filter(
          (clusterRoleBinding) => clusterRoleBinding.roleRef.name === 'cluster-admin',
      )
      .map((clusterRoleBinding) => ({
        name: clusterRoleBinding.metadata.name,
        role: clusterRoleBinding.roleRef.name,
        users: clusterRoleBinding.subjects
            .filter((sub) => sub.kind === 'User')
            .map((sub) => sub.name),
        groups: clusterRoleBinding.subjects
            .filter((sub) => sub.kind === 'Group')
            .map((sub) => sub.name),
      }));
}

async function findKymaAdminBindingForUser(targetUser) {
  const kymaAdminBindings = await getKymaAdminBindings();
  return kymaAdminBindings.find(
      (binding) => binding.users.indexOf(targetUser) >= 0,
  );
}

async function ensureKymaAdminBindingExistsForUser(targetUser) {
  const binding = await findKymaAdminBindingForUser(targetUser);
  expect(binding).not.to.be.undefined;
  expect(binding.users).to.include(targetUser);
}

async function ensureKymaAdminBindingDoesNotExistsForUser(targetUser) {
  const binding = await findKymaAdminBindingForUser(targetUser);
  expect(binding).to.be.undefined;
}

function getContainerStatusByImage(pod, image) {
  return pod.containerStatuses.find((status) => status.image === image);
}

const printRestartReport = (prevPodList = [], afterTestPodList = []) => {
  const report = prevPodList
      .map((elem) => {
      // check if the pod that existed before the test started still exists after test
        const afterTestPod = afterTestPodList.find(
            (arg) => arg.name === elem.name,
        );
        if (!afterTestPod || !afterTestPod.containerStatuses) {
          return {
            name: elem.name,
            containerRestarts: null,
          };
        }

        return {
          name: elem.name,
          containerRestarts: elem.containerStatuses
              .map((status) => {
                const afterTestContainerStatus = getContainerStatusByImage(
                    afterTestPod,
                    status.image,
                );

                let restartsTillTestStart;
                let message = '';
                if (!afterTestContainerStatus || !status) {
                  restartsTillTestStart = -1;
                  message = 'Container removed during report generation';
                } else {
                  restartsTillTestStart =
                afterTestContainerStatus.restartCount - status.restartCount;
                }

                return {
                  name: status.name,
                  image: status.image,
                  restartsTillTestStart: restartsTillTestStart,
                  info: message,
                };
              })
              .filter((status) => {
                // we're interested only in containers that crashed during test
                return status.restartsTillTestStart > 0;
              }),
        };
      })
      .filter((arg) => {
      // filter out pods that do not have statuses after test or somehow cannot be mapped to pods before test start
        return (
          Array.isArray(arg.containerRestarts) &&
        arg.containerRestarts.some(
            (container) => container.restartsTillTestStart !== 0,
        )
        );
      });
  if (report.length > 0) {
    console.log(`
=========RESTART REPORT========
${k8s.dumpYaml(report)}
===============================
`);
  }
};

let DEBUG = process.env.DEBUG === 'true';

function log(prefix, ...args) {
  if (args.length === 0) {
    return;
  }

  args = [...args];
  const fmt = `[${prefix}] ` + args[0];
  args = args.slice(1);
  console.log.apply(console, [fmt, ...args]);
}

function isDebugEnabled() {
  return DEBUG;
}

function switchDebug(on = true) {
  DEBUG = on;
}

function debug(...args) {
  if (!isDebugEnabled()) {
    return;
  }
  log('DEBUG', ...args);
}

function info(...args) {
  log('INFO', ...args);
}

function error(...args) {
  log('ERROR', ...args);
}

function fromBase64(s) {
  return Buffer.from(s, 'base64').toString('utf8');
}

function toBase64(s) {
  return Buffer.from(s).toString('base64');
}

function genRandom(len) {
  let res = '';
  const chrs = 'abcdefghijklmnopqrstuvwxyz0123456789';
  for (let i = 0; i < len; i++) {
    res += chrs.charAt(Math.floor(Math.random() * chrs.length));
  }

  return res;
}

function getEnvOrDefault(key, defValue = '') {
  if (!process.env[key]) {
    if (defValue !== '') {
      return defValue;
    }
    throw new Error(`Env ${key} not present`);
  }

  return process.env[key];
}

function getEnvOrThrow(key) {
  if (!process.env[key]) {
    throw new Error(`Env ${key} not present`);
  }

  return process.env[key];
}

function wait(fn, checkFn, timeout, interval) {
  return new Promise((resolve, reject) => {
    const th = setTimeout(function() {
      debug('wait timeout');
      done(reject, new Error('wait timeout'));
    }, timeout);
    const ih = setInterval(async function() {
      let res;
      try {
        res = await fn();
      } catch (ex) {
        res = ex;
      }
      checkFn(res) && done(resolve, res);
    }, interval);

    function done(fn, arg) {
      clearTimeout(th);
      clearInterval(ih);
      fn(arg);
    }
  });
}

async function patchApplicationGateway(name, ns) {
  const deployment = await retryPromise(
      async () => {
        return k8sAppsApi.readNamespacedDeployment(name, ns);
      },
      12,
      5000,
  ).catch(() => {
    throw new Error(`Timeout: ${name} is not ready`);
  });
  if (
    deployment.body.spec.template.spec.containers[0].args.includes(
        '--skipVerify=true',
    )
  ) {
    debug('Application Gateway already patched');
    return deployment;
  }

  const skipVerifyIndex =
    deployment.body.spec.template.spec.containers[0].args.findIndex((arg) =>
      arg.toString().includes('--skipVerify'),
    );
  expect(skipVerifyIndex).to.not.equal(-1);

  let replicaSets = await k8sAppsApi.listNamespacedReplicaSet(ns);
  const appGatewayRSsNames = replicaSets.body.items
      .filter((rs) => rs.metadata.labels['app'] === name)
      .map((r) => r.metadata.name);
  expect(appGatewayRSsNames.length).to.not.equal(0);

  const patch = [
    {
      op: 'replace',
      path: `/spec/template/spec/containers/0/args/${skipVerifyIndex}`,
      value: '--skipVerify=true',
    },
  ];
  const options = {
    headers: {'Content-type': k8s.PatchUtils.PATCH_FORMAT_JSON_PATCH},
  };
  await k8sAppsApi.patchNamespacedDeployment(
      name,
      ns,
      patch,
      undefined,
      undefined,
      undefined,
      undefined,
      options,
  );

  const patchedDeployment = await k8sAppsApi.readNamespacedDeployment(name, ns);
  expect(
      patchedDeployment.body.spec.template.spec.containers[0].args.findIndex(
          (arg) => arg.toString().includes('--skipVerify=true'),
      ),
  ).to.not.equal(-1);

  // We have to wait for the deployment to redeploy the actual pod.
  await sleep(1000);
  await waitForDeployment(name, ns);

  // Check if the new, patched pods are being created.
  // It's currently no k8s-js-native way to check if the new pods of
  // the deployment are running and the old ones are being terminated.
  replicaSets = await k8sAppsApi.listNamespacedReplicaSet(ns);
  const patchedAppGatewayRSs = replicaSets.body.items.filter(
      (rs) =>
        rs.metadata.labels['app'] === name &&
      !appGatewayRSsNames.includes(rs.metadata.name),
  );
  expect(patchedAppGatewayRSs.length).to.not.equal(0);
  await waitForReplicaSet(
      patchedAppGatewayRSs[0].metadata.name,
      ns,
      120 * 1000,
  );

  return patchedDeployment;
}

/**
 * Creates eventing subscription object that can be passed to the k8s API server
 * @param {string} eventType - full event type, e.g. sap.kyma.custom.commerce.order.created.v1
 * @param {string} sink URL where message should be dispatched eg. http://lastorder.test.svc.cluster.local
 * @param {string} name - subscription name
 * @param {string} namespace - namespace where subscription should be created
 * @return {object} JSON with subscription spec
 */
function eventingSubscription(eventType, sink, name, namespace) {
  return {
    apiVersion: 'eventing.kyma-project.io/v1alpha1',
    kind: 'Subscription',
    metadata: {
      name: `${name}`,
      namespace: namespace,
    },
    spec: {
      filter: {
        dialect: 'beb',
        filters: [
          {
            eventSource: {
              property: 'source',
              type: 'exact',
              value: '',
            },
            eventType: {
              property: 'type',
              type: 'exact',
              value: eventType /* sap.kyma.custom.commerce.order.created.v1*/,
            },
          },
        ],
      },
      sink: sink /* http://lastorder.test.svc.cluster.local*/,
    },
  };
}

function eventingSubscriptionV1Alpha2(eventType, source, sink, name, namespace, typeMatching='standard') {
  return {
    apiVersion: 'eventing.kyma-project.io/v1alpha2',
    kind: 'Subscription',
    metadata: {
      name: `${name}`,
      namespace: namespace,
    },
    spec: {
      source: source,
      typeMatching: typeMatching,
      types: [
        eventType,
      ],
      sink: sink /* http://lastorder.test.svc.cluster.local*/,
    },
  };
}

async function patchDeployment(name, ns, patch) {
  const options = {
    headers: {'Content-type': k8s.PatchUtils.PATCH_FORMAT_JSON_PATCH},
  };
  await k8sAppsApi.patchNamespacedDeployment(
      name,
      ns,
      patch,
      undefined,
      undefined,
      undefined,
      undefined,
      options,
  );
}

async function isKyma2() {
  try {
    const res = await k8sCoreV1Api.listNamespacedPod('kyma-installer');
    return res.body.items.length === 0;
  } catch (err) {
    throw new Error(`Error while trying to get pods in kyma-installer namespace: ${err.toString()}`);
  }
}

function namespaceObj(name) {
  return {
    apiVersion: 'v1',
    kind: 'Namespace',
    metadata: {name},
  };
}

/**
 * Creates eventing backend secret for event mesh (BEB)
 * @param {string} eventMeshSecretFilePath - file path of the EventMesh secret file
 * @param {string} name - name of the beb secret
 * @param {string} namespace - namespace where to create the secret
 * @return {json} - event mesh config data
 */
async function createEventingBackendK8sSecret(eventMeshSecretFilePath, name, namespace='default') {
  // read EventMesh secret from specified file
  const eventMeshSecret = JSON.parse(fs.readFileSync(eventMeshSecretFilePath, {encoding: 'utf8'}));

  const secretJson = {
    apiVersion: 'v1',
    kind: 'Secret',
    type: 'Opaque',
    metadata: {
      name,
      namespace,
    },
    data: {
      management: toBase64(JSON.stringify(eventMeshSecret['management'])),
      messaging: toBase64(JSON.stringify(eventMeshSecret['messaging'])),
      namespace: toBase64(eventMeshSecret['namespace']),
      serviceinstanceid: toBase64(eventMeshSecret['serviceinstanceid']),
      xsappname: toBase64(eventMeshSecret['xsappname']),
    },
  };

  // apply to k8s
  await k8sApply([secretJson], namespace, true);

  return {
    namespace: eventMeshSecret['namespace'],
    serviceinstanceid: eventMeshSecret['serviceinstanceid'],
    xsappname: eventMeshSecret['xsappname'],
  };
}

/**
 * Deletes eventing backend secret for event mesh (BEB)
 * @param {string} name - name of the beb secret
 * @param {string} namespace - namespace where the secret exists
 * @return {Promise<void>}
 */
function deleteEventingBackendK8sSecret(name, namespace='default') {
  const secretJson = {
    apiVersion: 'v1',
    kind: 'Secret',
    type: 'Opaque',
    metadata: {
      name,
      namespace,
    },
  };

  return k8sDelete([secretJson], namespace);
}

/**
 * Creates apirule for the service specified
 * @param {string} name - name of the configmap
 * @param {string} namespace - namespace where to create the configmap
 * @param {string} svcName - service to expose as apirule
 * @param {int} port - port of the service to expose as apirule
 */
async function createApiRuleForService(name, namespace='default', svcName, port) {
  const apiRuleJson = {
    apiVersion: 'gateway.kyma-project.io/v1beta1',
    kind: 'APIRule',
    metadata: {
      name,
      namespace,
    },
    spec: {
      gateway: 'kyma-gateway.kyma-system.svc.cluster.local',
      host: svcName,
      service: {
        name: svcName,
        port: port,
      },
      rules: [{
        accessStrategies: [{
          config: {},
          handler: 'allow',
        }],
        methods: ['GET'],
        path: '/.*',
      }],
    },
  };

  // apply to k8s
  await k8sApply([apiRuleJson], namespace, true);

  return waitForVirtualService(namespace, name);
}

/**
 * Deletes apirule
 * @param {string} name - name of the apirule
 * @param {string} namespace - namespace where the apirule exists
 * @return {Promise<void>}
 */
function deleteApiRule(name, namespace='default') {
  const apiRuleJson = {
    apiVersion: 'gateway.kyma-project.io/v1beta1',
    kind: 'APIRule',
    metadata: {
      name,
      namespace,
    },
  };

  return k8sDelete([apiRuleJson], namespace);
}

/**
 * Creates configmap with the data passed as argument
 * @param {Object} data - host name of the virtual service exposed to obtain the information
 * @param {string} name - name of the configmap
 * @param {string} namespace - namespace where to create the configmap
 */
async function createK8sConfigMap(data, name, namespace='default') {
  const configMapJson = {
    apiVersion: 'v1',
    kind: 'ConfigMap',
    metadata: {
      name,
      namespace,
    },
    data: data,
  };

  // apply to k8s
  await k8sApply([configMapJson], namespace, true);

  return waitForConfigMap(name, namespace);
}

/**
 * Deletes configmap
 * @param {string} name - name of the configmap
 * @param {string} namespace - namespace where the configmap exists
 * @return {Promise<void>}
 */
function deleteK8sConfigMap(name, namespace='default') {
  const configMapJson = {
    apiVersion: 'v1',
    kind: 'ConfigMap',
    metadata: {
      name,
      namespace,
    },
  };

  return k8sDelete([configMapJson], namespace);
}

/**
 * Switches eventing backend to specified backend-type (beb or nats)
 * @param {string} secretName - name of the beb secret
 * @param {string} namespace - namespace where the secret exists
 * @param {string} backendType - backend type to switch to. (beb or nats)
 */
async function switchEventingBackend(secretName, namespace='default', backendType='beb') {
  // patch data to label the eventing-backend secret
  const patch = [
    {
      'op': 'replace',
      'path': '/metadata/labels',
      'value': {
        'kyma-project.io/eventing-backend': backendType.toLowerCase(),
      },
    },
  ];

  const options = {'headers': {'Content-type': k8s.PatchUtils.PATCH_FORMAT_JSON_PATCH}};

  // apply k8s patch
  await k8sCoreV1Api.patchNamespacedSecret(
      secretName,
      namespace,
      patch,
      undefined,
      undefined,
      undefined,
      undefined,
      options,
  );

  await waitForEventingBackendToReady(backendType);
}

/**
 * Waits for eventing backend until its ready
 * @param {string} backendType - eventing backend type (beb or nats)
 * @param {string} name - name of the eventing backend
 * @param {string} namespace - namespace where the eventing backend exists
 * @param {number} timeout - timeout for waiting
 * @return {void}
 */
function waitForEventingBackendToReady(backendType='beb',
    name='eventing-backend',
    namespace = 'kyma-system',
    timeout = 180000) {
  return waitForK8sObject(
      `/apis/eventing.kyma-project.io/v1alpha1/namespaces/${namespace}/eventingbackends`,
      {},
      (_type, _apiObj, watchObj) => {
        return (
          watchObj.object.metadata.name === name &&
          watchObj.object.status.backendType.toLowerCase() === backendType.toLowerCase() &&
          watchObj.object.status.eventingReady === true
        );
      },
      timeout,
      `Waiting for eventing-backend: ${name} to get ready  timeout (${timeout} ms)`,
  );
}

/**
 * Prints logs of eventing-controller from kyma-system
 */
async function printEventingControllerLogs() {
  try {
    debug('Printing logs of eventing-controller from kyma-system');
    await printContainerLogs('app.kubernetes.io/name=controller,app.kubernetes.io/instance=eventing',
        'controller',
        'kyma-system',
        180000);
  } catch (err) {
    console.log(err);
    throw err;
  }
}

async function printStatusOfInClusterEventingInfrastructure(targetNamespace, encoding, funcName) {
  try {
    const kymaSystem = 'kyma-system';
    let publisherProxyReady = false;
    let eventingControllerReady = false;
    let natsServerReadyCounts = 0;
    let natsServerReady = false;
    let functionReady = false;

    const publisherDepl = await k8sAppsApi.listNamespacedDeployment(kymaSystem,
        undefined,
        undefined,
        undefined,
        undefined,
        'app.kubernetes.io/name=eventing-publisher-proxy, app.kubernetes.io/instance=eventing',
        undefined );
    if (publisherDepl.body.items[0].status.replicas === publisherDepl.body.items[0].status.readyReplicas) {
      publisherProxyReady = true;
    }

    const controllerDepl = await k8sAppsApi.listNamespacedDeployment(kymaSystem,
        undefined,
        undefined,
        undefined,
        undefined,
        'app.kubernetes.io/name=controller, app.kubernetes.io/instance=eventing',
        undefined );
    if (controllerDepl.body.items[0].status.replicas === controllerDepl.body.items[0].status.readyReplicas) {
      eventingControllerReady = true;
    }

    const natsServerPods = await k8sCoreV1Api.listNamespacedPod(kymaSystem,
        undefined,
        undefined,
        undefined,
        undefined,
        'kyma-project.io/dashboard=eventing, nats_cluster=eventing-nats');
    for (const nats of natsServerPods.body.items) {
      if (nats.status.phase === 'Running') {
        natsServerReadyCounts += 1;
      }
    }
    if (natsServerReadyCounts === natsServerPods.body.items.length) {
      natsServerReady = true;
    }

    const lastOrderFunc = await getFunction(funcName, targetNamespace);
    for (const cond of lastOrderFunc.status.conditions) {
      if (cond.type === 'Running' && cond.status === 'True') {
        functionReady = true;
        break;
      }
    }

    debug(`Printing status of infrastructure for in-cluster eventing, mode: ${encoding}`);
    debug(`Eventing-publisher-proxy deployment from ns: ${kymaSystem} ready: ${publisherProxyReady}`);
    debug(`Eventing-controller deployment from ns: ${kymaSystem} ready: ${eventingControllerReady}`);
    debug(`NATS-server pods from ns: ${kymaSystem} ready: ${natsServerReady}`);
    debug(`Function ${funcName} from ns: ${targetNamespace} ready: ${functionReady}`);
  } catch (err) {
    error(err);
    throw err;
  }
}

/**
 * Prints logs of eventing-publisher-proxy from kyma-system
 */
async function printEventingPublisherProxyLogs() {
  try {
    debug('Printing logs of eventing-publisher-proxy from kyma-system');
    await printContainerLogs('app.kubernetes.io/name=eventing-publisher-proxy, app.kubernetes.io/instance=eventing',
        'eventing-publisher-proxy',
        'kyma-system',
        180000);
  } catch (err) {
    error(err);
    throw err;
  }
}

async function printAllSubscriptions(testNamespace, crdVersion='v1alpha1') {
  try {
    debug(`Printing all subscriptions from namespace: ${testNamespace}`);
    const subs = await getAllSubscriptions(testNamespace, crdVersion);
    debug(JSON.stringify(subs, null, 4));
  } catch (err) {
    error(err);
    throw err;
  }
}

function waitForDeploymentWithLabel(
    labelKey,
    labelValue,
    namespace = 'default',
    timeout = 90000,
) {
  const query = {
    labelSelector: `${labelKey}=${labelValue}`,
  };
  return waitForK8sObject(
      `/apis/apps/v1/namespaces/${namespace}/deployments`,
      query,
      (_type, _apiObj, watchObj) => {
        return (

          watchObj.object.status.readyReplicas === 1
        );
      },
      timeout,
      `Waiting for deployment with label ${labelKey}=${labelValue} timeout (${timeout} ms)`,
  );
}

module.exports = {
  initializeK8sClient,
  getShootNameFromK8sServerUrl,
  retryPromise,
  convertAxiosError,
  ignore404,
  sleep,
  replaceAllInString,
  promiseAllSettled,
  kubectlApplyDir,
  kubectlApply,
  kubectlDelete,
  kubectlDeleteDir,
  k8sApply,
  k8sDelete,
  waitForK8sObject,
  waitForNamespace,
  waitForClusterAddonsConfiguration,
  waitForVirtualService,
  waitForDeployment,
  waitForService,
  waitForDaemonSet,
  waitForStatefulSet,
  waitForTokenRequest,
  waitForCompassConnection,
  waitForFunction,
  waitForSubscription,
  waitForPodWithLabel,
  waitForPodStatusWithLabel,
  waitForConfigMap,
  waitForSecret,
  waitForJob,
  deleteNamespaces,
  deleteAllK8sResources,
  getAllResourceTypes,
  getAllCRDs,
  getClusteraddonsconfigurations,
  ensureKymaAdminBindingExistsForUser,
  ensureKymaAdminBindingDoesNotExistsForUser,
  getSecret,
  getEventingBackend,
  getSecrets,
  getConfigMap,
  getSecretData,
  listResources,
  listResourceNames,
  k8sDynamicApi,
  k8sAppsApi,
  k8sCoreV1Api,
  getContainerRestartsForAllNamespaces,
  info,
  error,
  debug,
  switchDebug,
  isDebugEnabled,
  printRestartReport,
  fromBase64,
  toBase64,
  genRandom,
  getEnvOrThrow,
  wait,
  patchApplicationGateway,
  eventingSubscription,
  eventingSubscriptionV1Alpha2,
  getVirtualService,
  getGateway,
  getPersistentVolumeClaim,
  waitForApplicationCr,
  patchDeployment,
  isKyma2,
  namespaceObj,
  getEnvOrDefault,
  printContainerLogs,
  kubectlExecInPod,
  deleteK8sResource,
  deleteK8sObjects,
  deleteK8sPod,
  listPods,
  switchEventingBackend,
  waitForEventingBackendToReady,
  printAllSubscriptions,
  printEventingControllerLogs,
  printEventingPublisherProxyLogs,
  createEventingBackendK8sSecret,
  deleteEventingBackendK8sSecret,
  createK8sConfigMap,
  deleteK8sConfigMap,
  createApiRuleForService,
  deleteApiRule,
  printStatusOfInClusterEventingInfrastructure,
  getFunction,
  waitForEndpoint,
  waitForPodWithLabelAndCondition,
  waitForDeploymentWithLabel,
  getSubscription,
};
