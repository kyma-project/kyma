const {
  commerceMockTests,
  gettingStartedGuideTests,
} = require('./');

const {monitoringTests} = require('../monitoring');
const {loggingTests} = require('../logging');
const {tracingTests} = require('../tracing');

describe('Executing Standard Testsuite:', function() {
  commerceMockTests();
  gettingStartedGuideTests();

  monitoringTests();
  loggingTests();
  tracingTests();
});
