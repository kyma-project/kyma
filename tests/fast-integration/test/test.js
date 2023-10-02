
const {monitoringTests, unexposeGrafana} = require('../monitoring');

describe('Executing Standard Testsuite:', function() {
  this.timeout(10 * 60 * 1000);
  this.slow(5000);

  after('Unexpose Grafana', async function() {
    await unexposeGrafana();
  });

  monitoringTests();
});
