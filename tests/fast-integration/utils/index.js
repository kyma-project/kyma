const { expect } = require("chai");
const _ = require("lodash");
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

async function k8sApply(k8sDynamicApi, listOfSpecs, patch=true) {
  const options = { "headers": { "Content-type": 'application/merge-patch+json' } };
  return Promise.all(
    listOfSpecs.map(async (obj) => {
      try {
        const existing = await k8sDynamicApi.read(obj)
        if (patch) {
          obj.metadata.resourceVersion = existing.body.metadata.resourceVersion;
          return await k8sDynamicApi.patch(obj, undefined, undefined, undefined, undefined, options);
        } 
        return existing;
      }
      catch (e) {
        if (e.body && e.body.reason == 'NotFound') {
          return await k8sDynamicApi.create(obj)
        }
        else {
          console.log(e.body)
          throw e
        }
      }
    })
  );
}

function waitForK8sObject(watch, path, query, checkFn, timeout, timeoutMsg) {
  let res
  let timer
  const result = new Promise((resolve, reject) => {
    watch.watch(path, query, (type, apiObj, watchObj) => {
      if (checkFn(type, apiObj, watchObj)) {
        if (res) {
          res.abort();
        }
        clearTimeout(timer)
        resolve(watchObj.object)
      }
    }, () => { }).then((r) => { res = r; timer = setTimeout(() => { res.abort(); reject(new Error(timeoutMsg)) }, timeout); })
  });
  return result;
}

function waitForServiceClass(watch, name, namespace = "default") {
  return waitForK8sObject(watch, `/apis/servicecatalog.k8s.io/v1beta1/namespaces/${namespace}/serviceclasses`, {}, (_type, _apiObj, watchObj) => {
    return watchObj.object.spec.externalName.includes(name)
  }, 60 * 1000, `Waiting for ${name} service class timeout`);
}

function waitForServiceInstance(watch, name, namespace = "default") {
  return waitForK8sObject(watch, `/apis/servicecatalog.k8s.io/v1beta1/namespaces/${namespace}/serviceinstances`, {}, (_type, _apiObj, watchObj) => {
    return (watchObj.object.metadata.name == name && watchObj.object.status.conditions
      && watchObj.object.status.conditions.some((c) => (c.type == 'Ready' && c.status == 'True')))
  }, 60 * 1000, `Waiting for ${name} service instance timeout`);
}

function waitForServiceBinding(watch, name, namespace = "default") {
  return waitForK8sObject(watch, `/apis/servicecatalog.k8s.io/v1beta1/namespaces/${namespace}/servicebindings`, {}, (_type, _apiObj, watchObj) => {
    return (watchObj.object.metadata.name == name && watchObj.object.status.conditions
      && watchObj.object.status.conditions.some((c) => (c.type == 'Ready' && c.status == 'True')))
  }, 60 * 1000, `Waiting for ${name} service binding timeout`);
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
  waitForServiceBinding
};
