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

function shouldIgnoreAlert(alert) {
    var alertNamesToIgnore = [
        // Watchdog is an alert meant to ensure that the entire alerting pipeline is functional, so it should always be firing,
        "Watchdog",
        // Scrape limits can be exceeded on long-running clusters and can be ignored
        "ScrapeLimitForTargetExceeded",
    ]

    return alert.labels.severity == "critical" || alertNamesToIgnore.includes(alert.labels.alertname)
}

module.exports = {
    shouldIgnoreTarget,
    shouldIgnoreAlert,
};
