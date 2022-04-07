const {
  commerceMockTests,
  gettingStartedGuideTests,
} = require('./');

const {
  monitoringTests,
} = require('../monitoring');

describe('Executing Standard Testsuite:', function() {
  commerceMockTests();
  gettingStartedGuideTests();
  monitoringTests();
});
