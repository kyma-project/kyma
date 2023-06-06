const {
  KEBClient,
  KEBConfig,
} = require('../kyma-environment-broker');

const keb = new KEBClient(KEBConfig.fromEnv());
async function testEndpointWithoutToken(requestBody, endpoint, method) {
  try {
    await keb.callKEBWithoutToken(requestBody, endpoint, method);
  } catch (err) {
    throw new Error(`error while calling KEB endpoint "${endpoint}" without authorization: ${err.toString()}`);
  }
}

describe('KEB endpoints test', function() {
  const instanceID = 'keb-endpoints-test';
  const region = keb.getRegion();
  const testData = [
    {payload: {}, endpoint: `oauth/v2/service_instances/${instanceID}`, method: 'get'},
    {payload: {}, endpoint: `runtimes`, method: 'get'},
    {payload: {}, endpoint: `info/runtimes`, method: 'get'},
    {payload: {}, endpoint: `orchestrations`, method: 'get'},
    {payload: {}, endpoint: `oauth/${region}v2/service_instances/${instanceID}`, method: 'put'},
    {payload: {}, endpoint: `upgrade/cluster`, method: 'post'},
    {payload: {}, endpoint: `upgrade/kyma`, method: 'post'},
    {payload: {}, endpoint: `oauth/v2/service_instances/${instanceID}`, method: 'patch'},
    {payload: {}, endpoint: `oauth/v2/service_instances/${instanceID}`, method: 'delete'},
  ];

  it('Should reject call without authorization', async function() {
    for (const test of testData) {
      await testEndpointWithoutToken(test.payload, test.endpoint, test.method);
    }
  });
});

