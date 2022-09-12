const {
  timeoutTime,
  slowTime,
  cleanupTestingResources,
} = require('./utils');
const {printEventingControllerLogs} = require('../utils');

describe('Eventing tests cleanup', function() {
  this.timeout(timeoutTime);
  this.slow(slowTime);

  it('Printing the controller logs', async function() {
    await printEventingControllerLogs();
  });
  it('Cleaning: Test resources should be deleted', async function() {
    await cleanupTestingResources();
  });
});
