const fs = require('fs');
const eventMeshSecretFilePath = process.env.EVENTMESH_SECRET_FILE || '';
const natsBackend = 'nats';
const bebBackend = 'beb';
const kymaSystem = 'kyma-system';
const jaegerLabel = {
  key: 'app',
  value: 'jaeger',
};
const jaegerOperatorLabel = {
  key: 'app.kubernetes.io/name',
  value: 'tracing-jaeger-operator',
};

// returns the EventMesh namespace from the secret.
function getEventMeshNamespace() {
  try {
    if (eventMeshSecretFilePath === '') {
      return undefined;
    }
    const eventMeshSecret = JSON.parse(fs.readFileSync(eventMeshSecretFilePath, {encoding: 'utf8'}));
    return '/' + eventMeshSecret['namespace'];
  } catch (e) {
    console.error(e);
    return undefined;
  }
}

module.exports = {
  eventMeshSecretFilePath,
  getEventMeshNamespace,
  natsBackend,
  bebBackend,
  kymaSystem,
  jaegerLabel,
  jaegerOperatorLabel,
};
