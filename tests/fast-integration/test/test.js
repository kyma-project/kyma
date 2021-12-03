const {
    commerceMockTests,
    gettingStartedGuideTests,
    observabilityTests
} = require('./');

describe("Executing Standard Testsuite:", function() {
    commerceMockTests();
    gettingStartedGuideTests();
    observabilityTests();
});
