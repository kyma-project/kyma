const {
  timeoutTime,
  slowTime,
  cleanupTestingResources,
} = require('./utils');

describe('Eventing tests cleanup', function() {
  this.timeout(timeoutTime);
  this.slow(slowTime);

  it('Cleaning: Test namespaces should be deleted', async function() {
    await cleanupTestingResources();
  });
});
