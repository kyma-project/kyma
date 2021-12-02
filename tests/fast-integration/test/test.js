const {
    commerceMockTests,
    gettingStartedGuideTests,
    observabilityTests
} = require('./');

describe("Execute tests ->", function() {
    commerceMockTests();
    gettingStartedGuideTests();
    observabilityTests();
});
