#!/usr/bin/env groovy
import groovy.json.JsonSlurperClassic
/*

Monorepo root orchestrator: This Jenkinsfile runs the Jenkinsfiles of all subprojects based on the changes made and triggers kyma integration.
    - checks for changes since last successful build on master and compares to master if on a PR.
    - for every changed project, triggers related job async as configured in the seedjob.
    - for every changed additional project, triggers the kyma integration job.
    - passes info of:
        - revision
        - branch
        - current app version
        - docker registry directory to push images based on branch
        - all component versions

*/
def label = "kyma-${UUID.randomUUID().toString()}"
prRegex = /^PR-([0-9]+)$/ // PR format: PR-66
isMaster = env.BRANCH_NAME == 'master'
appVersion = ''
dockerPushRoot = ''

/*
    Projects that are built when changed, consisting of pairs [project path, produced docker image]. Projects producing multiple Docker images only need to provide one of them.
    Projects that do NOT produce docker images need to have a null value and will not be passed to the integration job, as there is nothing to deploy.

    IMPORTANT NOTE: Projects trigger jobs and therefore are expected to have a job defined with the same name.
*/
projects = [
    "docs": "kyma-docs",
    "components/api-controller": "api-controller",
    "components/apiserver-proxy": "apiserver-proxy",
    "components/binding-usage-controller": "binding-usage-controller",
    "components/configurations-generator": "configurations-generator",
    "components/environments": "environments",
    "components/istio-webhook": "istio-webhook",
    "components/helm-broker": "helm-broker",
    "components/remote-environment-broker": "remote-environment-broker",
    "components/remote-environment-controller": "remote-environment-controller",
    "components/metadata-service": "metadata-service",
    "components/gateway": "gateway",
    "components/installer": "installer",
    "components/connector-service": "connector-service",
    "components/ui-api-layer": "ui-api-layer",
    "components/event-bus": "event-bus-publish",
    "components/event-service": "event-service",
    "tools/alpine-net": "alpine-net",
    "tools/watch-pods": "watch-pods",
    "tools/stability-checker": "stability-checker",
    "tools/etcd-backup": "etcd-backup",
    "tools/etcd-tls-setup": "etcd-tls-setup",
    "tests/test-logging-monitoring": "test-logging-monitoring",
    "tests/logging": "test-logging",
    "tests/acceptance": "acceptance-tests",
    "tests/ui-api-layer-acceptance-tests": "ui-api-layer-acceptance-tests",
    "tests/gateway-tests": "gateway-acceptance-tests",
    "tests/test-environments": "test-environments",
    "tests/kubeless-test-client": "kubeless-test-client",
    "tests/api-controller-acceptance-tests": "api-controller-acceptance-tests",
    "tests/connector-service-tests": "connector-service-tests",
    "tests/metadata-service-tests": "metadata-service-tests",
    "tests/event-bus": "event-bus-e2e-tester",
    "governance": null
]

/*
    Projects that are NOT built when changed, but do trigger the kyma integration job.
*/
additionalProjects = ["resources","cluster","installation"]

/*
    project jobs to run are stored here to be sent into the parallel block outside the node executor.
*/
jobs = [:]

/*
    If true, Kyma integration will run at the end.
    NOTE: This is set automaticlly based on the changes detected.
*/
runIntegration = false

properties([
    buildDiscarder(logRotator(numToKeepStr: '10')),
    disableConcurrentBuilds()
])

podTemplate(label: label) {
    node(label) {
        try {
            timestamps {
                ansiColor('xterm') {
                    stage("setup") {
                        checkout scm
                        // use HEAD of branch as revision, Jenkins does a merge to master commit before starting this script, which will not be available on the jobs triggered below
                        commitID = sh (script: "git rev-parse origin/${env.BRANCH_NAME}", returnStdout: true).trim()
                        configureBuilds(commitID)
                        changes = changedProjects()

                        runIntegration = changes.size() > 0
                        if (changes.size() == 1 && changes[0] == "governance") {
                            runIntegration = false
                        }
                    }

                    stage('collect projects') {
                        buildableProjects = changes.intersect(projects.keySet()) // only projects that have build jobs
                        echo "Collected the following projects with changes: $buildableProjects..."
                        for (int i=0; i < buildableProjects.size(); i++) {
                            def index = i
                            jobs["${buildableProjects[index]}"] = { ->
                                    build job: "kyma/"+buildableProjects[index],
                                            wait: true,
                                            parameters: [
                                                string(name:'GIT_REVISION', value: "$commitID"),
                                                string(name:'GIT_BRANCH', value: "${env.BRANCH_NAME}"),
                                                string(name:'APP_VERSION', value: "$appVersion"),
                                                string(name:'PUSH_DIR', value: "$dockerPushRoot"),
                                                booleanParam(name:'FULL_BUILD', value: isMaster)
                                            ]
                            }
                        }
                    }
                }
            }
        } catch (ex) {
            echo "Got exception: ${ex}"
            currentBuild.result = "FAILURE"
            def body = "${currentBuild.currentResult} ${env.JOB_NAME}${env.BUILD_DISPLAY_NAME}: on branch: ${env.BRANCH_NAME}. See details: ${env.BUILD_URL}"
            emailext body: body, recipientProviders: [[$class: 'DevelopersRecipientProvider'], [$class: 'CulpritsRecipientProvider'], [$class: 'RequesterRecipientProvider']], subject: "${currentBuild.currentResult}: Job '${env.JOB_NAME} [${env.BUILD_NUMBER}]'"
        }
    }
}

// gather all component versions to pass to integration
if (runIntegration) {
    stage('collect versions') {
        versions = [:]
        
        changedProjects = jobs.keySet()
        for (int i = 0; i < changedProjects.size(); i++) {
            // only projects that have an associated docker image have a version to deploy
            if (projects["${changedProjects[i]}"] != null) {
                versions["${changedProjects[i]}"] = "$appVersion"
            }
        }

        unchangedProjects = projects.keySet() - changedProjects
        for (int i = 0; i < unchangedProjects.size(); i++) {
            // only projects that have an associated docker image have a version to deploy
            if (projects["${unchangedProjects[i]}"] != null) {
                versions["${unchangedProjects[i]}"] = projectVersion("${unchangedProjects[i]}")
            }
        }

        // convert versions to JSON string to pass on
        versions = versionsYaml(versions)
        echo """
    Component versions:
    $versions
    """
    }
}

// trigger jobs for projects that have changes, in parallel
stage('build projects') {
    parallel jobs
}

// trigger Kyma integration when changes are made to installation charts/code or resources
if (runIntegration) {
    stage('launch Kyma integration') {
        build job: 'kyma/integration',
            wait: true,
            parameters: [
                string(name:'GIT_REVISION', value: "$commitID"),
                string(name:'GIT_BRANCH', value: "${env.BRANCH_NAME}"),
                string(name:'APP_VERSION', value: "$appVersion"),
                string(name:'COMP_VERSIONS', value: "$versions") // YAML string
            ]
    }
}



/* -------- Helper Functions -------- */

/** Configure the parameters for the components to build:
 * - master: push root: "develop/" / image tag: APP_VERSION
 * - PR: push root: "pr/" / image tag: BRANCH NAME
 */
@NonCPS
def configureBuilds(commitID) {
    switch(env.BRANCH_NAME) {
        case "master": // master development builds
            dockerPushRoot = "develop/"
            appVersion = commitID.substring(0,8)
            break
        case ~prRegex: // PR changeset builds
            dockerPushRoot = "pr/"
            appVersion = env.BRANCH_NAME
            break
        default:
            error("Branch ${env.BRANCH_NAME} not supported in this pipeline.")
            break
    }
}

/**
 * Provides a list with the projects that have changes within the given projects list.
 * If no changes found, all projects will be returned.
 */
String[] changedProjects() {
    def res = []
    def projectPaths = projects.keySet()
    def allProjects = projectPaths + additionalProjects
    echo "Looking for changes in the following projects: $allProjects."

    // get all changes
    allChanges = changeset().split("\n")

    // if no changes build all projects
    if (allChanges.size() == 0) {
        echo "No changes found or could not be fetched, triggering all projects."
        return allProjects
    }

    // parse changeset and keep only relevant folders -> match with projects defined
    for (int i=0; i < allProjects.size(); i++) {
        for (int j=0; j < allChanges.size(); j++) {
            if (allChanges[j].startsWith(allProjects[i]) && changeIsValidFileType(allChanges[j],allProjects[i]) && !res.contains(allProjects[i])) {
                res.add(allProjects[i])
                break // already found a change in the current project, no need to continue iterating the changeset
            }
            if (allProjects[i] == "governance" && allChanges[j].endsWith(".md") && !res.contains(allProjects[i])) {
                res.add(allProjects[i])
                break // already found a change in one of the .md files, no need to continue iterating the changeset
            }
        }
    }

    return res
}

boolean changeIsValidFileType(String change, String project){
    return !change.endsWith(".md") || "docs".equals(project);
}

/**
 * Gets the changes on the Project based on the branch or an empty string if changes could not be fetched.
 */
@NonCPS
String changeset() {
    // on branch get changeset comparing with master
    if (env.BRANCH_NAME != "master") {
        echo "Fetching changes between origin/${env.BRANCH_NAME} and origin/master."
        return sh (script: "git --no-pager diff --name-only origin/master...origin/${env.BRANCH_NAME} | grep -v 'vendor\\|node_modules' || echo ''", returnStdout: true)
    }
    // on master get changeset since last successful commit
    else {
        echo "Fetching changes on master since last successful build."
        def successfulBuild = currentBuild.rawBuild.getPreviousSuccessfulBuild()
        if (successfulBuild) {
            def commit = commitHashForBuild(successfulBuild)
            return sh (script: "git --no-pager diff --name-only $commit 2> /dev/null | grep -v 'vendor\\|node_modules' || echo ''", returnStdout: true)
        }
    }
    return ""
}

/**
 * Gets the commit hash from a Jenkins build object
 */
@NonCPS
def commitHashForBuild(build) {
  def scmAction = build?.actions.find { action -> action instanceof jenkins.scm.api.SCMRevisionAction }
  return scmAction?.revision?.hash
}

/**
 * Fetches the newest released version of the given project from its manifest in the registry or an error if the version could not be fetched.
 * This function relies on the latest tag on docker images.
 * More info: https://docs.docker.com/registry/spec/manifest-v2-1/
 */
String projectVersion(project) {
    def img = projects[project]
    
    try {
        def json = "https://eu.gcr.io/v2/kyma-project/develop/${img}/manifests/latest".toURL().getText(requestProperties: [Accept: 'application/vnd.docker.distribution.manifest.v1+prettyjws'])
        def slurper = new JsonSlurperClassic()
        def doc = slurper.parseText(json)
        doc = slurper.parseText(doc.history[0].v1Compatibility)

        return doc.config.Labels.version

    } catch(e) {
        error("Error fetching latest version for ${project}: ${e}. Please check that ${project} has a docker image tagged ${img}:latest in the docker registry.\nLatest images are pushed to the registry on master branch builds.")
    }
}

/**
 * Provides the component versions as YAML; To be passed to the kyma installer in various jobs.
 */
@NonCPS
def versionsYaml(versions) {
    def overrides = 
"""
global.docs.version=${versions['docs']}
global.docs.dir=${versions['docs'] == env.BRANCH_NAME ? 'pr/' : 'develop/'}
global.apiserver_proxy.version=${versions['components/apiserver-proxy']}
global.apiserver_proxy.dir=${versions['components/apiserver-proxy'] == env.BRANCH_NAME ? 'pr/' : 'develop/'}
global.api_controller.version=${versions['components/api-controller']}
global.api_controller.dir=${versions['components/api-controller'] == env.BRANCH_NAME ? 'pr/' : 'develop/'}
global.binding_usage_controller.version=${versions['components/binding-usage-controller']}
global.binding_usage_controller.dir=${versions['components/binding-usage-controller'] == env.BRANCH_NAME ? 'pr/' : 'develop/'}
global.configurations_generator.version=${versions['components/configurations-generator']}
global.configurations_generator.dir=${versions['components/configurations-generator'] == env.BRANCH_NAME ? 'pr/' : 'develop/'}
global.environments.version=${versions['components/environments']}
global.environments.dir=${versions['components/environments'] == env.BRANCH_NAME ? 'pr/' : 'develop/'}
global.istio_webhook.version=${versions['components/istio-webhook']}
global.istio_webhook.dir=${versions['components/istio-webhook'] == env.BRANCH_NAME ? 'pr/' : 'develop/'}
global.helm_broker.version=${versions['components/helm-broker']}
global.helm_broker.dir=${versions['components/helm-broker'] == env.BRANCH_NAME ? 'pr/' : 'develop/'}
global.remote_environment_broker.version=${versions['components/remote-environment-broker']}
global.remote_environment_broker.dir=${versions['components/remote-environment-broker'] == env.BRANCH_NAME ? 'pr/' : 'develop/'}
global.metadata_service.version=${versions['components/metadata-service']}
global.metadata_service.dir=${versions['components/metadata-service'] == env.BRANCH_NAME ? 'pr/' : 'develop/'}
global.gateway.version=${versions['components/gateway']}
global.gateway.dir=${versions['components/gateway'] == env.BRANCH_NAME ? 'pr/' : 'develop/'}
global.installer.version=${versions['components/installer']}
global.installer.dir=${versions['components/installer'] == env.BRANCH_NAME ? 'pr/' : 'develop/'}
global.connector_service.version=${versions['components/connector-service']}
global.connector_service.dir=${versions['components/connector-service'] == env.BRANCH_NAME ? 'pr/' : 'develop/'}
global.ui_api_layer.version=${versions['components/ui-api-layer']}
global.ui_api_layer.dir=${versions['components/ui-api-layer'] == env.BRANCH_NAME ? 'pr/' : 'develop/'}
global.event_bus.version=${versions['components/event-bus']}
global.event_bus.dir=${versions['components/event-bus'] == env.BRANCH_NAME ? 'pr/' : 'develop/'}
global.event_service.version=${versions['components/event-service']}
global.event_service.dir=${versions['components/event-service'] == env.BRANCH_NAME ? 'pr/' : 'develop/'}
global.alpine_net.version=${versions['tools/alpine-net']}
global.alpine_net.dir=${versions['tools/alpine-net'] == env.BRANCH_NAME ? 'pr/' : 'develop/'}
global.watch_pods.version=${versions['tools/watch-pods']}
global.watch_pods.dir=${versions['tools/watch-pods'] == env.BRANCH_NAME ? 'pr/' : 'develop/'}
global.stability_checker.version=${versions['tools/stability-checker']}
global.stability_checker.dir=${versions['tools/stability-checker'] == env.BRANCH_NAME ? 'pr/' : 'develop/'}
global.etcd_backup.version=${versions['tools/etcd-backup']}
global.etcd_backup.dir=${versions['tools/etcd-backup'] == env.BRANCH_NAME ? 'pr/' : 'develop/'}
global.etcd_tls_setup.version=${versions['tools/etcd-tls-setup']}
global.etcd_tls_setup.dir=${versions['tools/etcd-tls-setup'] == env.BRANCH_NAME ? 'pr/' : 'develop/'}
global.test_logging_monitoring.version=${versions['tests/test-logging-monitoring']}
global.test_logging_monitoring.dir=${versions['tests/test-logging-monitoring'] == env.BRANCH_NAME ? 'pr/' : 'develop/'}
global.acceptance_tests.version=${versions['tests/acceptance']}
global.acceptance_tests.dir=${versions['tests/acceptance'] == env.BRANCH_NAME ? 'pr/' : 'develop/'}
global.ui_api_layer_acceptance_tests.version=${versions['tests/ui-api-layer-acceptance-tests']}
global.ui_api_layer_acceptance_tests.dir=${versions['tests/ui-api-layer-acceptance-tests'] == env.BRANCH_NAME ? 'pr/' : 'develop/'}
global.gateway_tests.version=${versions['tests/gateway-tests']}
global.gateway_tests.dir=${versions['tests/gateway-tests'] == env.BRANCH_NAME ? 'pr/' : 'develop/'}
global.test_environments.version=${versions['tests/test-environments']}
global.test_environments.dir=${versions['tests/test-environments'] == env.BRANCH_NAME ? 'pr/' : 'develop/'}
global.kubeless_test_client.version=${versions['tests/kubeless-test-client']}
global.kubeless_test_client.dir=${versions['tests/kubeless-test-client'] == env.BRANCH_NAME ? 'pr/' : 'develop/'}
global.api_controller_acceptance_tests.version=${versions['tests/api-controller-acceptance-tests']}
global.api_controller_acceptance_tests.dir=${versions['tests/api-controller-acceptance-tests'] == env.BRANCH_NAME ? 'pr/' : 'develop/'}
global.connector_service_tests.version=${versions['tests/connector-service-tests']}
global.connector_service_tests.dir=${versions['tests/connector-service-tests'] == env.BRANCH_NAME ? 'pr/' : 'develop/'}
global.metadata_service_tests.version=${versions['tests/metadata-service-tests']}
global.metadata_service_tests.dir=${versions['tests/metadata-service-tests'] == env.BRANCH_NAME ? 'pr/' : 'develop/'}
global.event_bus_tests.version=${versions['tests/event-bus']}
global.event_bus_tests.dir=${versions['tests/event-bus'] == env.BRANCH_NAME ? 'pr/' : 'develop/'}
global.test_logging.version=${versions['tests/logging']}
global.test_logging.dir=${versions['tests/logging'] == env.BRANCH_NAME ? 'pr/' : 'develop/'}
"""

    return "$overrides"
}
