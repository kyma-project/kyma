const {
  commerceMockTests,
  gettingStartedGuideTests,
} = require('./');

const {
  monitoringTests,
} = require('../monitoring');
const {
  loggingTests,
} = require('../logging');

describe('Executing Standard Testsuite:', function() {
  const testStartTimestamp = new Date().toISOString();
  commerceMockTests();
  gettingStartedGuideTests();
  monitoringTests();
  loggingTests(testStartTimestamp);
});
