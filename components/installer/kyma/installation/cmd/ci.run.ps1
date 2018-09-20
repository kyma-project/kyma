param (
    [switch]$IGNORE_TEST_FAIL = $false,
    [switch]$SKIP_TESTS = $false,
    [switch]$NON_INTERACTIVE = $false
)

$FINAL_IMAGE = "kyma-on-minikube"


$DOCKER_RUN_COMMAND = "docker run --rm -v /var/lib/docker"`
    + " -p 443:443"`
    + " -p 8443:8443"`
    + " -p 8001:8001"`
    + " -p 9411:9411"`
    + " -p 32000:32000"`
    + " -p 32001:32001"`
    + " --privileged"

if ($IGNORE_TEST_FAIL -eq $false) {
    $DOCKER_RUN_COMMAND = "${DOCKER_RUN_COMMAND} -e IGNORE_TEST_FAIL=false"
}

if ($SKIP_TESTS -eq $true) {
    $DOCKER_RUN_COMMAND = "${DOCKER_RUN_COMMAND} -e RUN_TESTS=false"
}

if ($NON_INTERACTIVE -eq $false) {
    $DOCKER_RUN_COMMAND = "${DOCKER_RUN_COMMAND} -it"
}

$DOCKER_RUN_COMMAND = "${DOCKER_RUN_COMMAND} ${FINAL_IMAGE}"

Write-Output "Running command: ${DOCKER_RUN_COMMAND}"

Invoke-Expression -Command $DOCKER_RUN_COMMAND