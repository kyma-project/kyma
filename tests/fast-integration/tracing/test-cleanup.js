const {
  timeoutTime,
  slowTime,
  cleanupTestingResources,
} = require('./utils');

async function testCleanup() {
  describe('Eventing tests cleanup', function() {
    this.timeout(timeoutTime);
    this.slow(slowTime);

    it('Cleaning: Test resources should be deleted', async function() {
      await cleanupTestingResources();
    });
  });
}
module.exports = {
  testCleanup,
};
