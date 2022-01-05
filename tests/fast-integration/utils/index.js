const stream = require('stream');
const k8s = require("@kubernetes/client-node");
const net = require("net");
const fs = require("fs");
const { join } = require("path");
const { expect } = require("chai");
const execa = require("execa");

const kc = new k8s.KubeConfig();
var k8sDynamicApi;
var k8sAppsApi;
var k8sCoreV1Api;
var k8sLog;
var k8sServerUrl;

var watch;
var forward;

function initializeK8sClient(opts) {
  opts = opts || {};
  try {
    if (opts.kubeconfigPath) {
      kc.loadFromFile(opts.kubeconfigPath);
    } else if (opts.kubeconfig) {
      kc.loadFromString(opts.kubeconfig);
    } else {
      kc.loadFromDefault();
    }

    k8sDynamicApi = kc.makeApiClient(k8s.KubernetesObjectApi);
    k8sAppsApi = kc.makeApiClient(k8s.AppsV1Api);
    k8sCoreV1Api = kc.makeApiClient(k8s.CoreV1Api);
    k8sRbacAuthorizationV1Api = kc.makeApiClient(k8s.RbacAuthorizationV1Api);
    k8sLog = new k8s.Log(kc);
    watch = new k8s.Watch(kc);
    forward = new k8s.PortForward(kc);
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
 * @returns {String}
 */
function getShootNameFromK8sServerUrl() {
  if (!k8sServerUrl || k8sServerUrl === "" || k8sServerUrl.split(".").length < 1) {
    throw new Error(`failed to get shootName from K8s server Url: ${k8sServerUrl}`)
  }
  return k8sServerUrl.split(".")[1];
}


/**
 * Retries a promise {retriesLeft} times every {interval} miliseconds
 *
 * @async
 * @param {function() : Promise} fn - async function that returns a promise
 * @param {number=} retriesLeft
 * @param {number=} interval
 * @throws
 * @returns {Promise}
 */
async function retryPromise(fn, retriesLeft = 10, interval = 30) {
  if (retriesLeft < 1) {
    throw new Error("retriesLeft argument should be greater then 0");
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
    message += ": " + JSON.stringify(axiosError.response.data);
  }
  return new Error(message);
}

const updateNamespacedResource =
  (client, group, version, pluralName) => async (name, namespace, updateFn) => {
    const obj = await client.getNamespacedCustomObject(
      group,
      version,
      namespace,
      pluralName,
      name
    );

    const updatedObj = updateFn(_.cloneDeep(obj.body));

    await client.replaceNamespacedCustomObject(
      group,
      version,
      namespace,
      pluralName,
      name,
      updatedObj
    );
  };

async function removeServiceInstanceFinalizer(client, name, namespace) {
  const serviceInstanceUpdater = updateNamespacedResource(
    client,
    "servicecatalog.k8s.io",
    "v1beta1",
    "serviceinstances"
  );

  await serviceInstanceUpdater(name, namespace, (obj) => {
    obj.metadata.finalizers = [];
    return obj;
  });
}

async function removeServiceBindingFinalizer(client, name, namespace) {
  const serviceBindingUpdater = updateNamespacedResource(
    client,
    "servicecatalog.k8s.io",
    "v1beta1",
    "servicebindings"
  );

  await serviceBindingUpdater(name, namespace, (obj) => {
    obj.metadata.finalizers = [];
    return obj;
  });
}

// "polyfill" for Promise.allSettled
async function promiseAllSettled(promises) {
  return Promise.all(
    promises.map((promise, i) =>
      promise
        .then((value) => ({
          status: "fulfilled",
          value,
        }))
        .catch((reason) => ({
          status: "rejected",
          reason,
        }))
    )
  );
}

/**
 *
 * @param {string} dir path to the directory with yaml files
 */
async function kubectlApplyDir(dir, namespace) {
  const files = fs.readdirSync(dir);
  for (let file of files) {
    if (file.endsWith(".yaml") || file.endsWith(".yml")) {
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
  for (let file of files) {
    if (file.endsWith(".yaml") || file.endsWith(".yml")) {
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

function kubectlPortForward(namespace, podName, port) {
  const server = net.createServer(function (socket) {
    forward.portForward(namespace, podName, [port], socket, null, socket, 3);
  });

  server.listen(port, "localhost");

  return () => {
    server.close();
  };
}

async function k8sDelete(listOfSpecs, namespace) {
  for (let res of listOfSpecs) {
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
          method: "DELETE",
        });
      } else {
        throw Error(
          "Object kind or metadata.selfLink is required to delete the resource"
        );
      }
      if (res.kind == "CustomResourceDefinition") {
        const version = res.spec.version || res.spec.versions[0].name;
        const path = `/apis/${res.spec.group}/${version}/${res.spec.names.plural}`;
        await deleteAllK8sResources(path);
      }
    } catch (err) {
      if (err.response.statusCode != 404) {
        throw err;
      }
    }
  }
}

async function getAllCRDs() {
  const path = "/apis/apiextensions.k8s.io/v1/customresourcedefinitions";
  const response = await k8sDynamicApi.requestPromise({
    url: k8sDynamicApi.basePath + path,
  });
  const stat = {};
  const body = JSON.parse(response.body);
  body.items.forEach(
    (crd) =>
      (stat[crd.spec.group] = stat[crd.spec.group]
        ? stat[crd.spec.group] + 1
        : 1)
  );
  return body.items;
}

async function getClusteraddonsconfigurations() {
  const path =
    "/apis/addons.kyma-project.io/v1alpha1/clusteraddonsconfigurations";
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

async function getPodPresets(namespace) {
  const path = `/apis/settings.svcat.k8s.io/v1alpha1/namespaces/${namespace}/podpresets/`;
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
  const body = JSON.parse(response.body);
  return body;
}

async function getFunction(name, namespace) {
  const path = `/apis/serverless.kyma-project.io/v1alpha1/namespaces/${namespace}/functions/${name}`;
  const response = await k8sDynamicApi.requestPromise({
    url: k8sDynamicApi.basePath + path,
  });
  const body = JSON.parse(response.body);
  return body;
}

async function getConfigMap(name, namespace) {
  const path = `/api/v1/namespaces/${namespace}/configmaps/${name}`;
  const response = await k8sDynamicApi.requestPromise({
    url: k8sDynamicApi.basePath + path,
  });
  const body = JSON.parse(response.body);
  return body;
}

async function getServiceInstance(name, namespace) {
  const path = `/apis/servicecatalog.k8s.io/v1beta1/namespaces/${namespace}/serviceinstances/${name}`;
  const response = await k8sDynamicApi.requestPromise({
    url: k8sDynamicApi.basePath + path,
  });
  const body = JSON.parse(response.body);
  return body;
}

async function k8sApply(resources, namespace, patch = true) {
  const options = {
    headers: { "Content-type": "application/merge-patch+json" },
  };
  for (let resource of resources) {
    if (!resource || !resource.kind || !resource.metadata.name) {
      debug("Skipping invalid resource:", resource);
      continue;
    }
    if (!resource.metadata.namespace) {
      resource.metadata.namespace = namespace;
    }
    try {
      const existing = await k8sDynamicApi.read(resource);
      await k8sDynamicApi.patch(
        resource,
        undefined,
        undefined,
        undefined,
        undefined,
        options
      );
      debug(resource.kind, resource.metadata.name, "reconfigured");
    } catch (e) {
      {
        if (e.body && e.body.reason == "NotFound") {
          try {
            await k8sDynamicApi.create(resource);
            debug(resource.kind, resource.metadata.name, "created");
          } catch (createError) {
            debug(resource.kind, resource.metadata.name, "failed to create");
            console.log(createError);
            throw createError;
          }
        } else {
          throw e;
        }
      }
    }
  }
}

function waitForK8sObject(path, query, checkFn, timeout, timeoutMsg) {
  debug("waiting for", path);
  return new Promise((resolve, reject) => {
    let onFulfilled = (req) => {
      setTimeout(() => {
        req.abort();
        reject(new Error(timeoutMsg));
      }, timeout);
    };
    watch.watch(
            path,
            query,
            (type, apiObj, watchObj) => {
              if (checkFn(type, apiObj, watchObj)) {
                debug("finished waiting for ", path);
                resolve(watchObj.object);
              }
            },
            (err) => {
              if (err) {
                reject(new Error(err))
              }
            }
        )
        .then(onFulfilled, reject).catch((reason) => {
      reject(new Error(reason));
    });
  });
}

function waitForClusterAddonsConfiguration(name, timeout = 90000) {
  return waitForK8sObject(
    `/apis/addons.kyma-project.io/v1alpha1/clusteraddonsconfigurations`,
    {},
    (_type, _apiObj, watchObj) => {
      return watchObj.object.metadata.name == name;
    },
    timeout,
    `Waiting for ${name} ClusterAddonsConfiguration timeout (${timeout} ms)`
  );
}

function waitForFunction(name, namespace = "default", timeout = 90000) {
  return waitForK8sObject(
    `/apis/serverless.kyma-project.io/v1alpha1/namespaces/${namespace}/functions`,
    {},
    (_type, _apiObj, watchObj) => {
      return (
        watchObj.object.metadata.name == name &&
        watchObj.object.status.conditions &&
        watchObj.object.status.conditions.some(
          (c) => c.type == "Running" && c.status == "True"
        )
      );
    },
    timeout,
    `Waiting for ${name} function timeout (${timeout} ms)`
  );
}

async function getAllSubscriptions(namespace = "default") {
  try {
    const path = `/apis/eventing.kyma-project.io/v1alpha1/namespaces/${namespace}/subscriptions`;
    const response = await k8sDynamicApi.requestPromise({
      url: k8sDynamicApi.basePath + path,
      qs: { limit: 500 },
    });
    const body = JSON.parse(response.body);

    return Promise.all(
      body.items.map((sub) => {
        return {
          apiVersion: sub["apiVersion"],
          spec: sub["spec"],
          status: sub["status"],

        }
      })
    ).then((results) => {
      return results.flat();
    });

  } catch (e) {
    if (e.statusCode == 404 || e.statusCode == 405) {
      // do nothing
    } else {
      console.log("Error:", e);
      throw e;
    }
  }
}

function waitForSubscription(name, namespace = "default", timeout = 180000) {
  return waitForK8sObject(
    `/apis/eventing.kyma-project.io/v1alpha1/namespaces/${namespace}/subscriptions`,
    {},
    (_type, _apiObj, watchObj) => {
      return (
        watchObj.object.metadata.name == name &&
        watchObj.object.status.conditions &&
        watchObj.object.status.conditions.some(
          (c) => c.type == "Subscription active" && c.status == "True"
        )
      );
    },
    timeout,
    `Waiting for ${name} subscription timeout (${timeout} ms)`
  );
}

function waitForClusterServiceBroker(name, timeout = 90000) {
  return waitForK8sObject(
      `/apis/servicecatalog.k8s.io/v1beta1/clusterservicebrokers`,
      {},
      (_type, _apiObj, watchObj) => {
        return (
            watchObj.object.metadata.name.includes(name) &&
            watchObj.object.status.conditions.some(
                (c) => c.type == "Ready" && c.status == "True"
            )
        );
      },
      timeout,
      `Waiting for ${name} cluster service broker (${timeout} ms)`
  );
}

function waitForServiceClass(name, namespace = "default", timeout = 90000) {
  return waitForK8sObject(
    `/apis/servicecatalog.k8s.io/v1beta1/namespaces/${namespace}/serviceclasses`,
    {},
    (_type, _apiObj, watchObj) => {
      return watchObj.object.spec.externalName.includes(name);
    },
    timeout,
    `Waiting for ${name} service class timeout (${timeout} ms)`
  );
}

function waitForServicePlanByServiceClass(
  serviceClassName,
  namespace = "default",
  timeout = 90000
) {
  return waitForK8sObject(
    `/apis/servicecatalog.k8s.io/v1beta1/namespaces/${namespace}/serviceplans`,
    {},
    (_type, _apiObj, watchObj) => {
      return watchObj.object.spec.serviceClassRef.name.includes(
        serviceClassName
      );
    },
    timeout,
    `Waiting for ${serviceClassName} service plan timeout (${timeout} ms)`
  );
}

function waitForServiceInstance(name, namespace = "default", timeout = 90000) {
  return waitForK8sObject(
    `/apis/servicecatalog.k8s.io/v1beta1/namespaces/${namespace}/serviceinstances`,
    {},
    (_type, _apiObj, watchObj) => {
      return (
        watchObj.object.metadata.name == name &&
        watchObj.object.status.conditions &&
        watchObj.object.status.conditions.some(
          (c) => c.type == "Ready" && c.status == "True"
        )
      );
    },
    timeout,
    `Waiting for service instance ${name} timeout (${timeout} ms)`
  );
}

function waitForServiceBinding(name, namespace = "default", timeout = 90000) {
  return waitForK8sObject(
    `/apis/servicecatalog.k8s.io/v1beta1/namespaces/${namespace}/servicebindings`,
    {},
    (_type, _apiObj, watchObj) => {
      return (
        watchObj.object.metadata.name == name &&
        watchObj.object.status.conditions &&
        watchObj.object.status.conditions.some(
          (c) => c.type == "Ready" && c.status == "True"
        )
      );
    },
    timeout,
    `Waiting for service binding ${name} timeout (${timeout} ms)`
  );
}

function waitForServiceBindingUsage(
  name,
  namespace = "default",
  timeout = 90000
) {
  return waitForK8sObject(
    `/apis/servicecatalog.kyma-project.io/v1alpha1/namespaces/${namespace}/servicebindingusages`,
    {},
    (_type, _apiObj, watchObj) => {
      return (
        watchObj.object.metadata.name == name &&
        watchObj.object.status.conditions &&
        watchObj.object.status.conditions.some(
          (c) => c.type == "Ready" && c.status == "True"
        )
      );
    },
    timeout,
    `Waiting for service binding usage ${name} timeout (${timeout} ms)`
  );
}

function waitForReplicaSet(name, namespace = "default", timeout = 90000) {
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
    `Waiting for replica set ${name} timeout (${timeout} ms)`
  );
}

function waitForDaemonSet(name, namespace = "default", timeout = 90000) {
  return waitForK8sObject(
    `/apis/apps/v1/watch/namespaces/${namespace}/daemonsets/${name}`,
    {},
    (_type, watchObj, _) => {
      return (
        watchObj.status.numberReady === watchObj.status.desiredNumberScheduled
      );
    },
    timeout,
    `Waiting for daemonset ${name} timeout (${timeout} ms)`
  );
}

function waitForDeployment(name, namespace = "default", timeout = 90000) {
  return waitForK8sObject(
    `/apis/apps/v1/namespaces/${namespace}/deployments`,
    {},
    (_type, _apiObj, watchObj) => {
      return (
        watchObj.object.metadata.name === name &&
        watchObj.object.status.conditions &&
        watchObj.object.status.conditions.some(
          (c) => c.type === "Available" && c.status === "True"
        )
      );
    },
    timeout,
    `Waiting for deployment ${name} timeout (${timeout} ms)`
  );
}

function waitForStatefulSet(name, namespace = "default", timeout = 90000) {
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
    `Waiting for StatefulSet ${name} timeout (${timeout} ms)`
  );
}

function waitForJob(name, namespace = "default", timeout = 900000, success = 1) {
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
    `Waiting for Job ${name} to succeed ${success} timeout (${timeout} ms)`
  );
}

async function kubectlExecInPod(pod, container, cmd, namespace = "default") {
  let execCmd = [`exec`, pod, `-c`, container, `-n`, namespace, '--', ...cmd];
  try {
    let out = await execa(`kubectl`, execCmd);
  } catch (error) {
    if (error.stdout === undefined) {
      throw error;
    }
    throw new Error(`failed to execute kubectl ${execCmd.join(" ")}:\n${error.stdout},\n${error.stderr}`);
  }
}

async function listPods(labelSelector, namespace = "default") {
  return await k8sCoreV1Api.listNamespacedPod(namespace, undefined, undefined, undefined, undefined, labelSelector);
}

async function printContainerLogs(labelSelector, container, namespace = "default", timeout = 90000) {
  const res = await k8sCoreV1Api.listNamespacedPod(namespace, undefined, undefined, undefined, undefined, labelSelector);
  res.body.items.sort((a,b) => {return a.metadata.creationTimestamp - b.metadata.creationTimestamp});
  for (const p of res.body.items) {
    process.stdout.write(`Getting logs for pod ${p.metadata.name}/${container}\n`);
    const logStream = new stream.PassThrough();
    logStream.on('data', (chunk) => {
        // use write rather than console.log to prevent double line feed
        process.stdout.write(chunk);
    });
    let end = new Promise(function(resolve, reject){
      logStream.on('end', () => {process.stdout.write("\n"); resolve()});
      logStream.on('error', reject);
    });
    k8sLog.log(namespace, p.metadata.name, container, logStream)
    await end;
  }
  process.stdout.write(`Done getting logs\n`);
}

function waitForVirtualService(namespace, apiRuleName, timeout = 20000) {
  const path = `/apis/networking.istio.io/v1beta1/namespaces/${namespace}/virtualservices`;
  const query = {
    labelSelector: `apirule.gateway.kyma-project.io/v1alpha1=${apiRuleName}.${namespace}`,
  };
  return waitForK8sObject(
    path,
    query,
    (_type, _apiObj, watchObj) => {
      return (
        watchObj.object.spec.hosts && watchObj.object.spec.hosts.length == 1
      );
    },
    timeout,
    `Wait for VirtualService ${apiRuleName} timeout (${timeout} ms)`
  );
}

async function getVirtualService(namespace, name) {
  const path = `/apis/networking.istio.io/v1beta1/namespaces/${namespace}/virtualservices/${name}`;
  const response = await k8sDynamicApi.requestPromise({
    url: k8sDynamicApi.basePath + path
  });
  const body = JSON.parse(response.body);
  return body.spec.hosts[0]
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
        watchObj.object.status.state == "OK" &&
        watchObj.object.status.url
      );
    },
    timeout,
    "Wait for TokenRequest timeout"
  );
}

function waitForCompassConnection(name, timeout = 90000) {
  const path = `/apis/compass.kyma-project.io/v1alpha1/compassconnections`;
  return waitForK8sObject(
    path,
    {},
    (_type, _apiObj, watchObj) => {
      return (
        watchObj.object.metadata.name === name &&
        watchObj.object.status.connectionState &&
        ["Connected", "Synchronized"].indexOf(
          watchObj.object.status.connectionState
        ) !== -1
      );
    },
    timeout,
    `Wait for Compass connection ${name} timeout (${timeout} ms)`
  );
}

function waitForPodWithLabel(
  labelKey,
  labelValue,
  namespace = "default",
  timeout = 90000
) {
  const query = {
    labelSelector: `${labelKey}=${labelValue}`,
  };
  return waitForK8sObject(
    `/api/v1/namespaces/${namespace}/pods`,
    query,
    (_type, _apiObj, watchObj) => {
      return (
        watchObj.object.status.phase == "Running" &&
        watchObj.object.status.containerStatuses.every((cs) => cs.ready)
      );
    },
    timeout,
    `Waiting for pod with label ${labelKey}=${labelValue} timeout (${timeout} ms)`
  );
}

function waitForConfigMap(
    cmName,
    namespace = "default",
    timeout = 90000
) {
  return waitForK8sObject(
      `/api/v1/namespaces/${namespace}/configmaps`,
      {},
      (_type, _apiObj, watchObj) => {
        return watchObj.object.metadata.name.includes(
            cmName
        );
      },
      timeout,
      `Waiting for ${cmName} service plan timeout (${timeout} ms)`
  );
}

async function deleteNamespaces(namespaces, wait = true) {
  let result = await k8sCoreV1Api.listNamespace();
  let allNamespaces = result.body.items.map((i) => i.metadata.name);
  namespaces = namespaces.filter((n) => allNamespaces.includes(n));
  if (namespaces.length == 0) {
    return;
  }
  const waitForNamespacesResult = waitForK8sObject(
    "/api/v1/namespaces",
    {},
    (type, apiObj, watchObj) => {
      if (type == "DELETED") {
        namespaces = namespaces.filter(
          (n) => n != watchObj.object.metadata.name
        );
      }
      return namespaces.length == 0 || !wait;
    },
    120 * 1000,
    "Timeout for deleting namespaces: " + namespaces
  );

  for (let name of namespaces) {
    k8sDynamicApi
      .delete({
        apiVersion: "v1",
        kind: "Namespace",
        metadata: { name },
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
    if (e.statusCode != 404 && e.statusCode != 405) {
      console.error("Error:", e);
      throw e;
    }
  }
  return [];
}

async function listResourceNames(path) {
  let resources = await listResources(path);
  return resources.map((o) => o.metadata.name);
}

async function resourceTypes(group, version) {
  const path = group ? `/apis/${group}/${version}` : `/api/${version}`;
  try {
    const response = await k8sDynamicApi.requestPromise({
      url: k8sDynamicApi.basePath + path,
      qs: { limit: 500 },
    });
    const body = JSON.parse(response.body);

    return body.resources.map((res) => {
      return { group, version, path, ...res };
    });
  } catch (e) {
    if (e.statusCode != 404 && e.statusCode != 405) {
      console.log("Error:", e);
      throw e;
    }
    return [];
  }
}

async function getAllResourceTypes() {
  try {
    const path = "/apis/apiregistration.k8s.io/v1/apiservices";
    const response = await k8sDynamicApi.requestPromise({
      url: k8sDynamicApi.basePath + path,
      qs: { limit: 500 },
    });
    const body = JSON.parse(response.body);
    return Promise.all(
      body.items.map((api) => {
        return resourceTypes(api.spec.group, api.spec.version);
      })
    ).then((results) => {
      return results.flat();
    });
  } catch (e) {
    if (e.statusCode == 404 || e.statusCode == 405) {
      // do nothing
    } else {
      console.log("Error:", e);
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
        const buff = Buffer.from(value, "base64");
        const decoded = buff.toString("ascii");
        return [key, decoded];
      })
    );
  } catch (e) {
    console.log("Error:", e);
    throw e;
  }
}

function ignore404(e) {
  if (e.statusCode != 404) {
    throw e;
  }
}

// NOTE: this no longer works, it relies on kube-api sending `selfLink` but the field has been deprecated
async function deleteAllK8sResources(
  path,
  query = {},
  retries = 2,
  interval = 1000,
  keepFinalizer = false
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
        for (let o of body.items) {
          deleteK8sResource(o, keepFinalizer);
        }
      } else if (!body.items) {
        deleteK8sResource(body, keepFinalizer);
      }
    }
  } catch (e) {
    debug("Error during delete ", path, String(e).substring(0, 1000));
  }
}

async function deleteK8sPod(o) {
  return await k8sCoreV1Api.deleteNamespacedPod(o.metadata.name, o.metadata.namespace);
}

// NOTE: this no longer works, it relies on kube-api sending `selfLink` but the field has been deprecated
async function deleteK8sResource(o, keepFinalizer = false) {
  if (o.metadata.finalizers && o.metadata.finalizers.length && !keepFinalizer) {
    const options = {
      headers: { "Content-type": "application/merge-patch+json" },
    };

    const obj = {
      kind: o.kind || "Secret", // Secret list doesn't return kind and apiVersion
      apiVersion: o.apiVersion || "v1",
      metadata: {
        name: o.metadata.name,
        namespace: o.metadata.namespace,
        finalizers: [],
      },
    };

    debug("Removing finalizers from", obj);

    await k8sDynamicApi
      .patch(obj, undefined, undefined, undefined, undefined, options)
      .catch(ignore404);
  }

  await k8sDynamicApi
    .requestPromise({
      url: k8sDynamicApi.basePath + o.metadata.selfLink,
      method: "DELETE",
    })
    .catch(ignore404);
  debug(
    "Deleted resource:",
    o.metadata.name,
    "namespace:",
    o.metadata.namespace
  );
}

async function getContainerRestartsForAllNamespaces() {
  const { body } = await k8sCoreV1Api.listPodForAllNamespaces();
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
  const { body } = await k8sRbacAuthorizationV1Api.listClusterRoleBinding();
  const adminRoleBindings = body.items;
  return adminRoleBindings
    .filter(
      (clusterRoleBinding) => clusterRoleBinding.roleRef.name === "cluster-admin"
    )
    .map((clusterRoleBinding) => ({
      name: clusterRoleBinding.metadata.name,
      role: clusterRoleBinding.roleRef.name,
      users: clusterRoleBinding.subjects
        .filter((sub) => sub.kind == "User")
        .map((sub) => sub.name),
      groups: clusterRoleBinding.subjects
        .filter((sub) => sub.kind == "Group")
        .map((sub) => sub.name),
    }));
}

async function findKymaAdminBindingForUser(targetUser) {
  let kymaAdminBindings = await getKymaAdminBindings();
  return kymaAdminBindings.find(
    (binding) => binding.users.indexOf(targetUser) >= 0
  );
}

async function ensureKymaAdminBindingExistsForUser(targetUser) {
  let binding = await findKymaAdminBindingForUser(targetUser);
  expect(binding).not.to.be.undefined;
  expect(binding.users).to.include(targetUser);
}

async function ensureKymaAdminBindingDoesNotExistsForUser(targetUser) {
  let binding = await findKymaAdminBindingForUser(targetUser);
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
        (arg) => arg.name === elem.name
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
              status.image
            );

            let restartsTillTestStart = 0;
            let message = "";
            if (!afterTestContainerStatus || !status) {
              restartsTillTestStart = -1;
              message = "Container removed during report generation";
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
          (container) => container.restartsTillTestStart != 0
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

function ignoreNotFound(e) {
  if (e.body && e.body.reason == "NotFound") {
    return;
  } else {
    console.log(e.body);
    throw e;
  }
}

let DEBUG = process.env.DEBUG;

function debug() {
  if (DEBUG) {
    console.log.apply(null, arguments);
  }
}

function switchDebug(on = true) {
  DEBUG = on;
}

function fromBase64(s) {
  return Buffer.from(s, "base64").toString("utf8");
}

function toBase64(s) {
  return Buffer.from(s).toString("base64");
}

function genRandom(len) {
  let res = "";
  const chrs = "abcdefghijklmnopqrstuvwxyz0123456789";
  for (let i = 0; i < len; i++) {
    res += chrs.charAt(Math.floor(Math.random() * chrs.length));
  }

  return res;
}

function getEnvOrDefault(key, defValue = "") {
  if (!process.env[key]) {
    if (defValue != "") {
      return defValue
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
    const th = setTimeout(function () {
      debug("wait timeout");
      done(reject, new Error("wait timeout"));
    }, timeout);
    const ih = setInterval(async function () {
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

async function ensureApplicationMapping(name, ns) {
  const applicationMapping = {
    apiVersion: "applicationconnector.kyma-project.io/v1alpha1",
    kind: "ApplicationMapping",
    metadata: { name: name, namespace: ns },
  };
  await k8sDynamicApi.delete(applicationMapping).catch(() => {});
  return await k8sDynamicApi.create(applicationMapping).catch((ex) => {
    debug(ex);
    throw ex;
  });
}

async function patchApplicationGateway(name, ns) {
  const deployment = await retryPromise(
    async () => {
      return k8sAppsApi.readNamespacedDeployment(name, ns);
    },
    12,
    5000
  ).catch((err) => {
    throw new Error(`Timeout: ${name} is not ready`);
  });
  if (
    deployment.body.spec.template.spec.containers[0].args.includes(
      "--skipVerify=true"
    )
  ) {
    debug("Application Gateway already patched");
    return deployment;
  }

  const skipVerifyIndex =
    deployment.body.spec.template.spec.containers[0].args.findIndex((arg) =>
      arg.toString().includes("--skipVerify")
    );
  expect(skipVerifyIndex).to.not.equal(-1);

  let replicaSets = await k8sAppsApi.listNamespacedReplicaSet(ns);
  const appGatewayRSsNames = replicaSets.body.items
    .filter((rs) => rs.metadata.labels["app"] === name)
    .map((r) => r.metadata.name);
  expect(appGatewayRSsNames.length).to.not.equal(0);

  const patch = [
    {
      op: "replace",
      path: `/spec/template/spec/containers/0/args/${skipVerifyIndex}`,
      value: "--skipVerify=true",
    },
  ];
  const options = {
    headers: { "Content-type": k8s.PatchUtils.PATCH_FORMAT_JSON_PATCH },
  };
  await k8sAppsApi.patchNamespacedDeployment(
    name,
    ns,
    patch,
    undefined,
    undefined,
    undefined,
    undefined,
    options
  );

  const patchedDeployment = await k8sAppsApi.readNamespacedDeployment(name, ns);
  expect(
    patchedDeployment.body.spec.template.spec.containers[0].args.findIndex(
      (arg) => arg.toString().includes("--skipVerify=true")
    )
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
      rs.metadata.labels["app"] === name &&
      !appGatewayRSsNames.includes(rs.metadata.name)
  );
  expect(patchedAppGatewayRSs.length).to.not.equal(0);
  await waitForReplicaSet(
    patchedAppGatewayRSs[0].metadata.name,
    ns,
    120 * 1000
  );

  return patchedDeployment;
}

/**
 * Creates eventing subscription object that can be passed to the k8s API server
 * @param {string} eventType - full event type, e.g. sap.kyma.custom.commerce.order.created.v1
 * @param {string} sink URL where message should be dispatched eg. http://lastorder.test.svc.cluster.local
 * @param {string} name - subscription name
 * @param {string} namespace - namespace where subscription should be created
 * @returns JSON with subscription spec
 */
function eventingSubscription(eventType, sink, name, namespace) {
  return {
    apiVersion: "eventing.kyma-project.io/v1alpha1",
    kind: "Subscription",
    metadata: {
      name: `${name}`,
      namespace: namespace,
    },
    spec: {
      filter: {
        dialect: "beb",
        filters: [
          {
            eventSource: {
              property: "source",
              type: "exact",
              value: "",
            },
            eventType: {
              property: "type",
              type: "exact",
              value: eventType /*sap.kyma.custom.commerce.order.created.v1*/,
            },
          },
        ],
      },
      protocol: "BEB",
      protocolsettings: {
        exemptHandshake: true,
        qos: "AT-LEAST-ONCE",
      },
      sink: sink /*http://lastorder.test.svc.cluster.local*/,
    },
  };
}

async function patchDeployment(name, ns, patch) {
  const options = {
    headers: { "Content-type": k8s.PatchUtils.PATCH_FORMAT_JSON_PATCH },
  };
  await k8sAppsApi.patchNamespacedDeployment(
    name,
    ns,
    patch,
    undefined,
    undefined,
    undefined,
    undefined,
    options
  );
}

async function isKyma2() {
  try {
    const res = await k8sCoreV1Api.listNamespacedPod("kyma-installer");
    return res.body.items.length === 0;
  } catch(err) {
    throw new Error(`Error while trying to get pods in kyma-installer namespace: ${err.toString()}`);
  }
}

function namespaceObj(name) {
  return {
    apiVersion: "v1",
    kind: "Namespace",
    metadata: { name },
  };
}

function serviceInstanceObj(name, serviceClassExternalName) {
  return {
    apiVersion: "servicecatalog.k8s.io/v1beta1",
    kind: "ServiceInstance",
    metadata: {
      name: name,
    },
    spec: { serviceClassExternalName },
  };
}

function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

/**
 * Creates eventing backend secret for event mesh (BEB)
 * @param {string} eventMeshSecretFilePath - file path of the EventMesh secret file
 * @param {string} name - name of the beb secret
 * @param {string} namespace - namespace where to create the secret
 * @returns {json} - event mesh config data
 */
 async function createEventingBackendK8sSecret(eventMeshSecretFilePath, name, namespace="default") {
  // read EventMesh secret from specified file
  const eventMeshSecret = JSON.parse(fs.readFileSync(eventMeshSecretFilePath, { encoding: "utf8" }))

  const secretJson = {
    apiVersion: "v1",
    kind: "Secret",
    type: "Opaque",
    metadata: {
      name,
      namespace
    },
    data: { 
      management: toBase64(JSON.stringify(eventMeshSecret["management"])),
      messaging: toBase64(JSON.stringify(eventMeshSecret["messaging"])),
      namespace: toBase64(eventMeshSecret["namespace"]),
      serviceinstanceid: toBase64(eventMeshSecret["serviceinstanceid"]),
      xsappname: toBase64(eventMeshSecret["xsappname"])
    },
  };

  // apply to k8s
  await k8sApply([secretJson], namespace, true);

  return {
    namespace: eventMeshSecret["namespace"],
    serviceinstanceid: eventMeshSecret["serviceinstanceid"],
    xsappname: eventMeshSecret["xsappname"]
  }
}

/**
 * Deletes eventing backend secret for event mesh (BEB)
 * @param {string} name - name of the beb secret
 * @param {string} namespace - namespace where the secret exists
 * @returns
 */
 function deleteEventingBackendK8sSecret(name, namespace="default") {
  const secretJson = {
    apiVersion: "v1",
    kind: "Secret",
    type: "Opaque",
    metadata: {
      name,
      namespace
    }
  };

  return k8sDelete([secretJson], namespace)
 }

/**
 * Switches eventing backend to specified backend-type (beb or nats)
 * @param {string} secretName - name of the beb secret
 * @param {string} namespace - namespace where the secret exists
 * @param {string} backendType - backend type to switch to. (beb or nats)
 * @returns
 */
async function switchEventingBackend(secretName, namespace="default", backendType="beb") {
  // patch data to label the eventing-backend secret
  const patch = [
    {
        "op": "replace",
        "path":"/metadata/labels",
        "value": {
          "kyma-project.io/eventing-backend": backendType.toLowerCase()
        }
    }
  ];

  const options = { "headers": { "Content-type": k8s.PatchUtils.PATCH_FORMAT_JSON_PATCH}};

  // apply k8s patch
  await k8sCoreV1Api.patchNamespacedSecret(
    secretName,
    namespace,
    patch,
    undefined,
    undefined,
    undefined,
    undefined,
    options
  );

  await sleep(30 * 1000); // Putting on sleep because there may be a delay in eventing-backend status update propagation
  await waitForEventingBackendToReady(backendType)
}

/**
 * Waits for eventing backend until its ready
 * @param {string} name - name of the eventing backend
 * @param {string} namespace - namespace where the eventing backend exists
 * @param {string} backendType - eventing backend type (beb or nats)
 * @returns
 */
function waitForEventingBackendToReady(backendType="beb", name="eventing-backend", namespace = "kyma-system", timeout = 180000) {
  return waitForK8sObject(
    `/apis/eventing.kyma-project.io/v1alpha1/namespaces/${namespace}/eventingbackends`,
    {},
    (_type, _apiObj, watchObj) => {
      return (
        watchObj.object.metadata.name == name &&
        watchObj.object.status.backendType.toLowerCase() == backendType.toLowerCase() &&
        watchObj.object.status.eventingReady == true &&
        watchObj.object.status.publisherProxyReady == true &&
        watchObj.object.status.subscriptionControllerReady == true
      );
    },
    timeout,
    `Waiting for eventing-backend: ${name} to get ready  timeout (${timeout} ms)`
  );
}

/**
 * Prints logs of eventing-controller from kyma-system
 */
async function printEventingControllerLogs() {
  try{
    console.log(`****** Printing logs of eventing-controller from kyma-system`)
    await printContainerLogs('app.kubernetes.io/name=controller, app.kubernetes.io/instance=eventing', "controller", 'kyma-system', 180000);
  }
  catch(err) {
    console.log(err)
    throw err
  }
}

/**
 * Prints status of the components on which in-cluster eventing happens
 */
 async function printStatusOfInClusterEventingInfrastructure(targetNamespace, encoding, funcName) {
  try{
    let kymaSystem = "kyma-system"
    let publisherProxyReady = false
    let eventingControllerReady = false
    let natsServerReadyCounts = 0
    let natsServerReady = false
    let functionReady = false

    let publisherDeployment = await k8sAppsApi.listNamespacedDeployment(kymaSystem,undefined, undefined,undefined, undefined, 'app.kubernetes.io/name=eventing-publisher-proxy, app.kubernetes.io/instance=eventing', undefined );
    if (publisherDeployment.body.items[0].status.replicas === publisherDeployment.body.items[0].status.readyReplicas) {
      publisherProxyReady = true
    }

    let controllerDeployment  = await k8sAppsApi.listNamespacedDeployment(kymaSystem,undefined, undefined,undefined, undefined, 'app.kubernetes.io/name=controller, app.kubernetes.io/instance=eventing', undefined );
    if (controllerDeployment.body.items[0].status.replicas === controllerDeployment.body.items[0].status.readyReplicas) {
      eventingControllerReady = true
    }

    let natsServerPods  = await k8sCoreV1Api.listNamespacedPod(kymaSystem, undefined, undefined, undefined, undefined, 'kyma-project.io/dashboard=eventing, nats_cluster=eventing-nats');
    for (let nats of natsServerPods.body.items) {
      if (nats.status.phase === "Running") {
        natsServerReadyCounts += 1
      }
    }
    if (natsServerReadyCounts === natsServerPods.body.items.length) {
      natsServerReady = true
    }

    let lastOrderFunc = await getFunction(funcName, targetNamespace);
    for (let cond of lastOrderFunc.status.conditions) {
      if (cond.type === "Running" && cond.status === "True") {
        functionReady = true
        break
      }
    }

    console.log(`****** Printing status of infrastructure for in-cluster eventing, mode: ${encoding} *******`)
    console.log(`****** Eventing-publisher-proxy deployment from ns: ${kymaSystem} ready: ${publisherProxyReady}`)
    console.log(`****** Eventing-controller deployment from ns: ${kymaSystem} ready: ${eventingControllerReady}`)
    console.log(`****** NATS-server pods from ns: ${kymaSystem} ready: ${natsServerReady}`)
    console.log(`****** Function ${funcName} from ns: ${targetNamespace} ready: ${functionReady}`)
    console.log(`****** End *******`)
  }
  catch(err) {
    console.log(err)
    throw err
  }
}

/**
 * Prints logs of eventing-publisher-proxy from kyma-system
 */
 async function printEventingPublisherProxyLogs() {
  try{ 
    console.log(`****** Printing logs of eventing-publisher-proxy from kyma-system`)
    await printContainerLogs('app.kubernetes.io/name=eventing-publisher-proxy, app.kubernetes.io/instance=eventing', "eventing-publisher-proxy", 'kyma-system', 180000);
  }
  catch(err) {
    console.log(err)
    throw err
  }
}

/**
 * Prints subscriptions json
 */
 async function printAllSubscriptions(testNamespace) {
  try{
    console.log(`****** Printing all subscriptions from namespace: ${testNamespace}`)
    const subs = await getAllSubscriptions(testNamespace)
    console.log(JSON.stringify(subs, null, 4))
  }
  catch(err) {
    console.log(err)
    throw err
  }
}

// getTraceDAG returns a DAG for the provided Jaeger tracing data
async function getTraceDAG(trace) {
  // Find root spans which are not child of any other span
  const rootSpans = trace["spans"].filter((s) => !(s["references"].find((r) => r["refType"] === "CHILD_OF")))

  // Find and attach child spans for each root span
  for (const root of rootSpans) {
    await attachTraceChildSpans(root, trace);
  }
  return rootSpans
}

// attachChildSpans finds child spans of current parentSpan and attach it to parentSpan object
// and also recursively, finds and attaches further child spans of each child.
async function attachTraceChildSpans(parentSpan, trace) {
  // find child spans of current parentSpan and attach it to parentSpan object
  parentSpan["childSpans"] = trace["spans"].filter((s) => s["references"].find((r) => r["refType"] === "CHILD_OF" && r["spanID"] === parentSpan["spanID"] && r["traceID"] === parentSpan["traceID"]));
  // recursively, find and attach further child span of each parentSpan["childSpans"]
  if (parentSpan["childSpans"] && parentSpan["childSpans"].length > 0) {
    for (const child of parentSpan["childSpans"]) {
      await attachTraceChildSpans(child, trace);
    }
  }
}

module.exports = {
  initializeK8sClient,
  getShootNameFromK8sServerUrl,
  retryPromise,
  convertAxiosError,
  removeServiceInstanceFinalizer,
  removeServiceBindingFinalizer,
  sleep,
  promiseAllSettled,
  kubectlApplyDir,
  kubectlApply,
  kubectlDelete,
  kubectlDeleteDir,
  kubectlPortForward,
  k8sApply,
  k8sDelete,
  waitForK8sObject,
  waitForClusterAddonsConfiguration,
  waitForClusterServiceBroker,
  waitForServiceClass,
  waitForServicePlanByServiceClass,
  waitForServiceInstance,
  waitForServiceBinding,
  waitForServiceBindingUsage,
  waitForVirtualService,
  waitForDeployment,
  waitForDaemonSet,
  waitForStatefulSet,
  waitForTokenRequest,
  waitForCompassConnection,
  waitForFunction,
  waitForSubscription,
  waitForPodWithLabel,
  waitForConfigMap,
  waitForJob,
  deleteNamespaces,
  deleteAllK8sResources,
  getAllResourceTypes,
  getAllCRDs,
  getClusteraddonsconfigurations,
  ensureKymaAdminBindingExistsForUser,
  ensureKymaAdminBindingDoesNotExistsForUser,
  getSecret,
  getSecrets,
  getConfigMap,
  getPodPresets,
  getSecretData,
  getServiceInstance,
  listResources,
  listResourceNames,
  k8sDynamicApi,
  k8sAppsApi,
  k8sCoreV1Api,
  getContainerRestartsForAllNamespaces,
  debug,
  switchDebug,
  printRestartReport,
  fromBase64,
  toBase64,
  genRandom,
  getEnvOrThrow,
  wait,
  ensureApplicationMapping,
  patchApplicationGateway,
  eventingSubscription,
  getVirtualService,
  patchDeployment,
  isKyma2,
  namespaceObj,
  serviceInstanceObj,
  getEnvOrDefault,
  printContainerLogs,
  kubectlExecInPod,
  deleteK8sResource,
  deleteK8sPod,
  listPods,
  switchEventingBackend,
  waitForEventingBackendToReady,
  printAllSubscriptions,
  printEventingControllerLogs,
  printEventingPublisherProxyLogs,
  createEventingBackendK8sSecret,
  deleteEventingBackendK8sSecret,
  getTraceDAG,
  printStatusOfInClusterEventingInfrastructure,
};
