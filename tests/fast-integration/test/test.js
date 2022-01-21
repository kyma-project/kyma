const {
  commerceMockTests,
  gettingStartedGuideTests,
  monitoringTests,
} = require('./');

describe('Executing Standard Testsuite:', function() {
  commerceMockTests();
  gettingStartedGuideTests();
  monitoringTests();
});
