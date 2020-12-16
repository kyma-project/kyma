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
          console.log(`retriesLeft: ${retriesLeft}`);
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

async function waitForTokenRequestReady(
  client,
  name,
  namespace,
  retriesLeft = 10,
  interval = 3000
) {
  return await retryPromise(
    async () => {
      console.log("Trying to read TokenRequest .status");
      return client
        .getNamespacedCustomObject(
          "applicationconnector.kyma-project.io",
          "v1alpha1",
          namespace,
          "tokenrequests",
          name
        )
        .then((res) => {
          expect(res.body).to.have.nested.property("status.url");
          return res;
        });
    },
    retriesLeft,
    interval
  ).catch(expectNoK8sErr);
}

module.exports = {
  retryPromise,
  expectNoK8sErr,
  expectNoAxiosErr,
  removeServiceInstanceFinalizer,
  removeServiceBindingFinalizer,
  sleep,
  promiseAllSettled,
  waitForTokenRequestReady,
};
