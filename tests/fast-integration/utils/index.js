const k8s = require("@kubernetes/client-node");
const net = require("net")
const fs = require("fs");
const { join } = require("path");
const { expect } = require("chai");

const kc = new k8s.KubeConfig();
var k8sDynamicApi;
var k8sAppsApi;
var k8sCoreV1Api;

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
    watch = new k8s.Watch(kc);
    forward = new k8s.PortForward(kc);
  } catch (err) {
    console.log(err.message);
  }

}
initializeK8sClient();



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
    return axiosError;
  }
  if (axiosError.response && axiosError.response.status && axiosError.response.statusText) {
    message += `\n${axiosError.response.status}: ${axiosError.response.statusText}`;
  }
  if (axiosError.response && axiosError.response.data) {
    message += ": " + JSON.stringify(axiosError.response.data);
    debug(axiosError.response.data);
  }
  return new Error(message)
}

const updateNamespacedResource = (client, group, version, pluralName) => async (
  name,
  namespace,
  updateFn
) => {
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
  for (let file of files) {
    if (file.endsWith('.yaml') || file.endsWith('.yml')) {
      await kubectlDelete(join(dir, file), namespace).catch(console.error);
    }
  }
}

function kubectlDelete(file, namespace) {
  const yaml = fs.readFileSync(file);
  const listOfSpecs = k8s.loadAllYaml(yaml);
  return k8sDelete(listOfSpecs, namespace);
}

function kubectlPortForward(namespace, podName, port) {
  const server = net.createServer(function (socket) {
    forward.portForward(namespace, podName, [port], socket, null, socket, 3)
  })

  server.listen(port, 'localhost');

  return () => { server.close() }
}

async function k8sDelete(listOfSpecs, namespace) {
  for (let res of listOfSpecs) {
    if (namespace) {
      res.metadata.namespace = namespace;
    }
    debug(`Delete ${res.metadata.name}`);
    try {
      if (res.kind) {
        await k8sDynamicApi.delete(res)
      } else if (res.metadata.selfLink) {
        await k8sDynamicApi.requestPromise({
          url: k8sDynamicApi.basePath + res.metadata.selfLink,
          method: "DELETE",
        });
      } else {
        throw Error("Object kind or metadata.selfLink is required to delete the resource")
      }
      if (res.kind == "CustomResourceDefinition") {
        const version = res.spec.version || res.spec.versions[0].name;
        const path = `/apis/${res.spec.group}/${version}/${res.spec.names.plural}`
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
  const path = '/apis/apiextensions.k8s.io/v1/customresourcedefinitions';
  const response = await k8sDynamicApi.requestPromise({
    url: k8sDynamicApi.basePath + path
  });
  const stat = {}
  const body = JSON.parse(response.body);
  body.items.forEach(crd => stat[crd.spec.group] = stat[crd.spec.group] ? stat[crd.spec.group] + 1 : 1)
  return body.items

}

async function k8sApply(resources, namespace, patch = true) {
  const options = { "headers": { "Content-type": 'application/merge-patch+json' } };
  for (let resource of resources) {
    if (!resource || !resource.kind || !resource.metadata.name) {
      debug("Skipping invalid resource:", resource);
      continue;
    }
    if (!resource.metadata.namespace) {
      resource.metadata.namespace = namespace
    }
    try {
      const existing = await k8sDynamicApi.read(resource)
      await k8sDynamicApi.patch(resource, undefined, undefined, undefined, undefined, options);
      debug(resource.kind, resource.metadata.name, "reconfigured")
    } catch (e) {
      {
        if (e.body && e.body.reason == 'NotFound') {
          try {
            await k8sDynamicApi.create(resource);
            debug(resource.kind, resource.metadata.name, "created")
          } catch (createError) {
            debug(resource.kind, resource.metadata.name, "failed to create");
            console.log(createError)
            throw createError;
          }
        } else {
          throw e
        }
      }
    }

  }
}

function waitForK8sObject(path, query, checkFn, timeout, timeoutMsg) {
  debug("waiting for", path);
  let res;
  let timer;
  const result = new Promise((resolve, reject) => {
    watch
      .watch(
        path,
        query,
        (type, apiObj, watchObj) => {
          if (checkFn(type, apiObj, watchObj)) {
            if (res) {
              res.abort();
            }
            clearTimeout(timer);
            debug("finished waiting for ", path);
            resolve(watchObj.object);
          }
        },
        () => { }
      )
      .then((r) => {
        res = r;
        timer = setTimeout(() => {
          res.abort();
          reject(new Error(timeoutMsg));
        }, timeout);
      });
  });
  return result;
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

function waitForSubscription(name, namespace = "default", timeout = 90000) {
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
    `Waiting for ${name} function timeout (${timeout} ms)`
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

function waitForDeployment(name, namespace = "default", timeout = 90000) {
  return waitForK8sObject(
    `/apis/apps/v1/namespaces/${namespace}/deployments`,
    {},
    (_type, _apiObj, watchObj) => {
      return (
        watchObj.object.metadata.name == name &&
        watchObj.object.status.conditions &&
        watchObj.object.status.conditions.some(
          (c) => c.type == "Available" && c.status == "True"
        )
      );
    },
    timeout,
    `Waiting for deployment ${name} timeout (${timeout} ms)`
  );
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

function waitForTokenRequest(name, namespace, timeout = 5000) {
  const path = `/apis/applicationconnector.kyma-project.io/v1alpha1/namespaces/${namespace}/tokenrequests`;
  return waitForK8sObject(path, {}, (_type, _apiObj, watchObj) => {
    return watchObj.object.metadata.name === name && watchObj.object.status && watchObj.object.status.state == "OK"
      && watchObj.object.status.url;
  }, timeout, "Wait for TokenRequest timeout");
}

function waitForCompassConnection(name, timeout = 90000) {
  const path = `/apis/compass.kyma-project.io/v1alpha1/compassconnections`;
  return waitForK8sObject(path, {}, (_type, _apiObj, watchObj) => {
    return watchObj.object.metadata.name === name && watchObj.object.status.connectionState
      && ["Connected", "Synchronized"].indexOf(watchObj.object.status.connectionState) !== -1
  }, timeout, `Wait for Compass connection ${name} timeout (${timeout} ms)`);
}

function waitForPodWithLabel(labelKey, labelValue, namespace = "default", timeout = 90000) {
  const query = {
    labelSelector: `${labelKey}=${labelValue}`,
  };
  return waitForK8sObject(
    `/api/v1/namespaces/${namespace}/pods`,
    query,
    (_type, _apiObj, watchObj) => {
      return (
        watchObj.object.status.phase == "Running" &&
        watchObj.object.status.containerStatuses.every(cs => cs.ready)
      );
    },
    timeout,
    `Waiting for pod with label ${labelKey}=${labelValue} timeout (${timeout} ms)`
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

function ignore404(e) {
  if (e.statusCode != 404) {
    throw e;
  }
}

async function deleteAllK8sResources(
  path,
  query = {},
  retries = 2,
  interval = 1000
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
          deleteK8sResource(o);
        }
      } else if (!body.items) {
        deleteK8sResource(body);
      }
    }
  } catch (e) {
    debug("Error during delete ", path, String(e).substring(0, 1000));
  }
}

async function deleteK8sResource(o) {
  if (o.metadata.finalizers && o.metadata.finalizers.length) {
    const options = {
      headers: { "Content-type": "application/merge-patch+json" },
    };

    const obj = {
      kind: o.kind || "Secret", // Secret list doesn't return kind and apiVersion
      apiVersion: o.apiVersion || "v1",
      metadata: { name: o.metadata.name, namespace: o.metadata.namespace, finalizers: [] },
    };

    debug("Removing finalizers from", obj);

    await k8sDynamicApi
      .patch(obj, undefined, undefined, undefined, undefined, options)
      .catch(ignore404);
  }

  await k8sDynamicApi
    .requestPromise({ url: k8sDynamicApi.basePath + o.metadata.selfLink, method: 'DELETE' })
    .catch(ignore404);
  debug("Deleted resource:", o.metadata.name, 'namespace:', o.metadata.namespace);
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
            return {
              name: status.name,
              image: status.image,
              restartsTillTestStart:
                afterTestContainerStatus.restartCount - status.restartCount,
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
  return Buffer
    .from(s, "base64")
    .toString("utf8");
}

function toBase64(s) {
  return Buffer
    .from(s)
    .toString("base64");
}

function genRandom(len) {
  let res = "";
  const chrs = "abcdefghijklmnopqrstuvwxyz0123456789";
  for (let i = 0; i < len; i++) {
    res += chrs.charAt(Math.floor(Math.random() * chrs.length));
  }

  return res;
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
      try { res = await fn(); }
      catch (ex) { res = ex; }
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
    metadata: { name: name, namespace: ns }
  };
  await k8sDynamicApi.delete(applicationMapping).catch(() => { });
  return await k8sDynamicApi.create(applicationMapping).catch((ex) => { debug(ex); throw ex });
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

  const skipVerifyIndex = deployment.body.spec.template.spec.containers[0].args.findIndex(
      arg => arg.toString().includes('--skipVerify')
  );
  expect(skipVerifyIndex).to.not.equal(-1);

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
  await k8sDynamicApi
    .requestPromise({
      url:
        k8sDynamicApi.basePath +
        deployment.body.metadata.selfLink,
      method: "PATCH",
      body: patch,
      json: true,
      headers: options.headers,
    });

  const patchedDeployment = await k8sAppsApi.readNamespacedDeployment(name, ns);
  expect(patchedDeployment.body.spec.template.spec.containers[0].args.findIndex(
      arg => arg.toString().includes('--skipVerify=true')
  )).to.not.equal(-1);

  // We have to wait for the deployment to redeploy the actual pod.
  await sleep(1000);
  await waitForDeployment(name, ns);

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
        filters: [{
          eventSource: {
            property: "source", type: "exact", value: "",
          },
          eventType: {
            property: "type", type: "exact", value: eventType/*sap.kyma.custom.commerce.order.created.v1*/
          }
        }]
      },
      protocol: "BEB",
      protocolsettings: {
        exemptHandshake: true,
        qos: "AT-LEAST-ONCE",
      },
      sink: sink/*http://lastorder.test.svc.cluster.local*/
    }
  }
}

module.exports = {
  initializeK8sClient,
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
  waitForServiceClass,
  waitForServiceInstance,
  waitForServiceBinding,
  waitForServiceBindingUsage,
  waitForVirtualService,
  waitForDeployment,
  waitForTokenRequest,
  waitForCompassConnection,
  waitForFunction,
  waitForSubscription,
  waitForPodWithLabel,
  deleteNamespaces,
  deleteAllK8sResources,
  getAllResourceTypes,
  getAllCRDs,
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
};

