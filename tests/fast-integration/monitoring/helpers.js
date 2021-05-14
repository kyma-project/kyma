function shouldIgnoreTarget(target) {
    let podsToBeIgnored = [
        // Ignore the pods that are created during tests.
        "-testsuite-",
        "test",
        "nodejs12-",
        "nodejs14-",
        "upgrade",
        // Ignore the pods created by jobs which are executed after installation of control-plane.
        "compass-migration",
        "compass-director-tenant-loader-default",
        "compass-agent-configuration",
    ];

    let namespacesToBeIgnored = ["test", "e2e"];

    return podsToBeIgnored.includes(target.pod) || namespacesToBeIgnored.includes(target.namespace);
}

module.exports = {
    shouldIgnoreTarget
};
