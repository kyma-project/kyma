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
  commerceMockTests();
  gettingStartedGuideTests();
  monitoringTests();
  loggingTests();
});
