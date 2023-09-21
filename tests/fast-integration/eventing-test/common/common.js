const fs = require('fs');
const eventMeshSecretFilePath = process.env.EVENTMESH_SECRET_FILE || '';
const natsBackend = 'nats';
const bebBackend = 'beb';
const kymaSystem = 'kyma-system';
const conditionReady = {
  condition: 'Ready',
  status: 'True',
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
  conditionReady
};
