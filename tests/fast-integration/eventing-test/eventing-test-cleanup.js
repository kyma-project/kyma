const {
  timeoutTime,
  slowTime,
  cleanupTestingResources, isSKR,
} = require('./utils');
const {resetGrafanaProxy} = require('../monitoring');

describe('Eventing tests cleanup', function() {
  this.timeout(timeoutTime);
  this.slow(slowTime);

  it('Cleaning: Test resources should be deleted', async function() {
    await cleanupTestingResources();

    if (!isSKR) {
      await resetGrafanaProxy();
    }
  });
});
