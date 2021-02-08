const k8s = require("@kubernetes/client-node");
const { expect } = require("chai");
const _ = require("lodash");

const kc = new k8s.KubeConfig();
kc.loadFromDefault();

const k8sDynamicApi = kc.makeApiClient(k8s.KubernetesObjectApi);
const k8sAppsApi = kc.makeApiClient(k8s.AppsV1Api);
const k8sCoreV1Api = kc.makeApiClient(k8s.CoreV1Api);

const watch = new k8s.Watch(kc);

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

function expectNoK8sErr(err) {
  if (!!err.body && !!err.body.message) {
    expect(err.body.message).to.be.empty; // handle native k8s errors
  }

  expect(err).to.be.undefined; // handle rest of errors
}

function expectNoAxiosErr(err) {
  if (err.reponse) {
    // https://github.com/axios/axios#handling-errors
    const { request, ...errorObject } = err.response; // request is too verbose
    expect(errorObject).to.deep.eq({});
  }

  expect(err).to.be.undefined; // catch non-axios errors
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

async function k8sApply(listOfSpecs, namespace, patch = true) {
  const options = {
    headers: { "Content-type": "application/merge-patch+json" },
  };

  return Promise.all(
    listOfSpecs.map(async (obj) => {
      try {
        if (namespace) {
          obj.metadata.namespace = namespace;
        }
        debug(
          "k8sApply:",
          obj.metadata.name,
          ", kind:",
          obj.kind,
          ", apiVersion:",
          obj.apiVersion,
          ", namespace:",
          obj.metadata.namespace
        );
        const existing = await k8sDynamicApi.read(obj);

        if (patch) {
          obj.metadata.resourceVersion = existing.body.metadata.resourceVersion;
          return await k8sDynamicApi.patch(
            obj,
            undefined,
            undefined,
            undefined,
            undefined,
            options
          );
        }
        return existing;
      } catch (e) {
        if (e.body && e.body.reason == "NotFound") {
          return await k8sDynamicApi.create(obj).catch((e) => {
            throw new Error(
              "k8sApply error, cannot create: " + JSON.stringify(obj)
            );
          });
        } else {
          //console.log(e.body)
          throw e;
        }
      }
    })
  );
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
        () => {}
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
      return listObj.items.map((o) => o.metadata.name);
    }
  } catch (e) {
    if (e.statusCode != 404 && e.statusCode != 405) {
      console.error("Error:", e);
      throw e;
    }
  }
  return [];
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
  const options = {
    headers: { "Content-type": "application/merge-patch+json" },
  };
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
      if (body.items.length == 0) {
        break;
      }
      for (let o of body.items) {
        if (o.metadata.finalizers && o.metadata.finalizers.length) {
          const obj = {
            kind: o.kind,
            apiVersion: o.apiVersion,
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
    }
  } catch (e) {
    debug("Error during delete ", path, String(e).substring(0, 1000));
  }
}

async function getContainerRestartsForAllNamespaces() {
  const { body } = await k8sCoreV1Api.listPodForAllNamespaces();
  const pods = body.items;
  return pods
    .filter((pd) => !!pd.status && !!pd.status.containerStatuses)
    .map((pod) => ({
      name: pod.metadata.name,
      containerStatuses:
        pod.status &&
        (pod.status.containerStatuses || []).map((elem) => ({
          name: elem.name,
          image: elem.image,
          restartCount: elem.restartCount,
        })),
    }));
}

const getRestartRaport = (prevPodList, afterPodList) => {
  return prevPodList
    .map((elem) => {
      const afterPod = afterPodList.find((arg) => arg.name === elem.name);
      if (!afterPod || !afterPod.containerStatuses) {
        return {
          name: elem.name,
          containerRestarts: null,
        };
      }

      return {
        name: elem.name,
        containerRestarts: elem.containerStatuses
          .map((status) => {
            let afterContainerStatus = afterPod.containerStatuses.find(
              (pd) => status.image === pd.image
            );
            return {
              name: status.name,
              image: status.image,
              restartsTillTestStart:
                afterContainerStatus.restartCount - status.restartCount,
            };
          })
          .filter((status) => status.restartsTillTestStart > 0),
      };
    })
    .filter((arg) => !!arg.containerRestarts)
    .filter((arg) =>
      arg.containerRestarts.some(
        (container) => container.restartsTillTestStart != 0
      )
    );
};

function ignoreNotFound(e) {
  if (e.body && e.body.reason == "NotFound") {
  } else {
    console.log(e.body);
    throw e;
  }
}
const DEBUG = process.env.DEBUG;

function debug() {
  if (DEBUG) {
    console.log.apply(null, arguments);
  }
}

module.exports = {
  retryPromise,
  expectNoK8sErr,
  expectNoAxiosErr,
  removeServiceInstanceFinalizer,
  removeServiceBindingFinalizer,
  sleep,
  promiseAllSettled,
  k8sApply,
  waitForK8sObject,
  waitForServiceClass,
  waitForServiceInstance,
  waitForServiceBinding,
  waitForServiceBindingUsage,
  waitForVirtualService,
  waitForDeployment,
  deleteNamespaces,
  deleteAllK8sResources,
  getAllResourceTypes,
  listResources,
  k8sDynamicApi,
  k8sAppsApi,
  getContainerRestartsForAllNamespaces,
  debug,
  getRestartRaport,
};
