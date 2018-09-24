#!/usr/bin/env groovy
import groovy.json.JsonSlurperClassic
/*

Monorepo releaser: This Jenkinsfile runs the Jenkinsfiles of all subprojects based on the changes made and triggers kyma integration.
    - checks for changes since last successful build on master and compares to master if on a PR.
    - for every changed project, triggers related job async as configured in the seedjob.
    - for every changed additional project, triggers the kyma integration job.
    - passes info of:
        - revision
        - branch
        - current app version
        - all component versions

*/
def label = "kyma-${UUID.randomUUID().toString()}"
semVerRegex = /^([0-9]+\.[0-9]+\.[0-9]+)$/ // semVer format: 1.2.3
releaseBranchRegex = /^release\/([0-9]+\.[0-9]+)$/ // release branch format: release/1.5
isRelease = params.RELEASE_VERSION ==~ semVerRegex

commitID = ''
appVersion = ''
dockerPushRoot = ''

/*
    Projects that will be released.

    IMPORTANT NOTE: Projects trigger jobs and therefore are expected to have a job defined with the same name.
*/
projects = [
    "docs",
    "components/api-controller",
    "components/binding-usage-controller",
    "components/configurations-generator",
    "components/environments",
    "components/istio-webhook",
    "components/helm-broker",
    "components/remote-environment-broker",
    "components/remote-environment-controller",
    "components/metadata-service",
    "components/gateway",
    "components/installer",
    "components/connector-service",
    "components/ui-api-layer",
    "components/event-bus",
    "components/event-service",
    "tools/alpine-net",
    "tools/watch-pods",
    "tools/stability-checker",
    "tools/etcd-backup",
    "tools/etcd-tls-setup",
    "tests/test-logging-monitoring",
    "tests/logging",
    "tests/acceptance",
    "tests/ui-api-layer-acceptance-tests",
    "tests/gateway-tests",
    "tests/test-environments",
    "tests/kubeless-test-client",
    "tests/api-controller-acceptance-tests",
    "tests/connector-service-tests",
    "tests/metadata-service-tests",
    "tests/event-bus",
    "governance",
]

/*
    project jobs to run are stored here to be sent into the parallel block outside the node executor.
*/
jobs = [:]

properties([
    buildDiscarder(logRotator(numToKeepStr: '30')),
    disableConcurrentBuilds()
])

podTemplate(label: label) {
    node(label) {
        try {
            timestamps {
                ansiColor('xterm') {
                    stage("setup") {
                        checkout scm

                        // validate parameters
                        if (!isRelease && !params.RELEASE_VERSION.empty()) {
                            error("Release version ${params.RELEASE_VERSION} does not follow semantic versioning.")
                        }
                        if (!params.RELEASE_BRANCH ==~ releaseBranchRegex) {
                            error("Release branch ${params.RELEASE_BRANCH} is not a valid branch. Provide a branch such as 'release/0.5'")
                        }
                    
                        commitID = sh (script: "git rev-parse origin/${params.RELEASE_BRANCH}", returnStdout: true).trim()
                        configureBuilds(commitID)
                    }

                    stage('collect projects') {
                        buildableProjects = projects.keySet() // only projects that have build jobs
                        for (int i=0; i < buildableProjects.size(); i++) {
                            def index = i
                            jobs["${buildableProjects[index]}"] = { ->
                                    build job: "kyma/"+buildableProjects[index],
                                            wait: true,
                                            parameters: [
                                                string(name:'GIT_REVISION', value: "$commitID"),
                                                string(name:'GIT_BRANCH', value: "${params.RELEASE_BRANCH}"),
                                                string(name:'APP_VERSION', value: "$appVersion"),
                                                string(name:'PUSH_DIR', value: "$dockerPushRoot"),
                                                booleanParam(name:'FULL_BUILD', value: true)
                                            ]
                            }
                        }
                    }
                }
            }
        } catch (ex) {
            echo "Got exception: ${ex}"
            currentBuild.result = "FAILURE"
            def body = "${currentBuild.currentResult} ${env.JOB_NAME}${env.BUILD_DISPLAY_NAME}: on branch/tag: ${params.RELEASE_BRANCH}. See details: ${env.BUILD_URL}"
            emailext body: body, recipientProviders: [[$class: 'DevelopersRecipientProvider'], [$class: 'CulpritsRecipientProvider'], [$class: 'RequesterRecipientProvider']], subject: "${currentBuild.currentResult}: Job '${env.JOB_NAME} [${env.BUILD_NUMBER}]'"
        }
    }
}


// release components
stage('build projects') {
    parallel jobs
}

stage('launch Kyma integration') {
    build job: 'kyma/integration',
        wait: true,
        parameters: [
            string(name:'GIT_REVISION', value: "$commitID"),
            string(name:'GIT_BRANCH', value: "${params.RELEASE_BRANCH}"),
            string(name:'APP_VERSION', value: "$appVersion"),
            string(name:'COMP_VERSIONS', value: "${versionsYaml()}") // YAML string
        ]
}

stage("Publish ${isRelease ? 'Release' : 'Prerelease'} ${appVersion}") {
    def slurper = new JsonSlurperClassic()
    def zip = "${appVersion}.tar.gz"
    
    // create release zip
    writeFile file: "./installation/versions.env", text: "${versionsYaml()}"
    sh "tar -czf ${zip} ./installation ./resources"

    // create release on github
    withCredentials([string(credentialsId: 'public-github-token', variable: 'token')]) {
        // TODO add changelog as "body"
        def data = "'{\"tag_name\": \"${appVersion}\",\"target_commitish\": \"${commitID}\",\"name\": \"${appVersion}\",\"body\": \"Release ${appVersion}\",\"draft\": false,\"prerelease\": ${isRelease ? 'false' : 'true'}}'"
        def json = sh (script: "curl --data ${data} -H \"Authorization: token $token\" https://api.github.com/repos/kyma-project/kyma/releases", returnStdout: true)
        def releaseID = slurper.parseText(json).id

        // upload zip file
        sh "curl --data-binary @$zip -H \"Authorization: token $token\" -H \"Content-Type: application/zip\" https://uploads.github.com/repos/kyma-project/kyma/releases/${releaseID}/assets?name=${zip}"
        // upload versions env file
        sh "curl --data-binary @installation/versions.env -H \"Authorization: token $token\" -H \"Content-Type: text/plain\" https://uploads.github.com/repos/kyma-project/kyma/releases/${releaseID}/assets?name=${appVersion}.env"
    }
}

stage("Upload versions file to azure") {
    // TODO upload RC versions.env to azure
}




/* -------- Helper Functions -------- */

/** Configure the parameters for the components to build:
 * - release candidate: push root: "rc/" / image tag: short commit ID
 * - release: push root: "" / image tag: semantic version
 */
def configureBuilds(commitID) {
    if(isRelease) {
        echo ("Building Release ${params.RELEASE_VERSION}")
        dockerPushRoot = ""
        appVersion = params.RELEASE_VERSION
    } else {
        echo ("Building Release Candidate for ${params.RELEASE_BRANCH}")
        dockerPushRoot = "rc/"
        appVersion = (params.RELEASE_BRANCH =~ /([0-9]+\.[0-9]+)$/)[0][1] // release branch number (e.g. 1.0)
    }   
}

/**
 * Provides the component versions as YAML; To be passed to the kyma installer in various jobs.
 */
@NonCPS
def versionsYaml() {
    def overrides = 
"""
global.docs.version=${appVersion}
global.docs.dir=${dockerPushRoot}
global.api_controller.version=${appVersion}
global.api_controller.dir=${dockerPushRoot}
global.binding_usage_controller.version=${appVersion}
global.binding_usage_controller.dir=${dockerPushRoot}
global.configurations_generator.version=${appVersion}
global.configurations_generator.dir=${dockerPushRoot}
global.environments.version=${appVersion}
global.environments.dir=${dockerPushRoot}
global.istio_webhook.version=${appVersion}
global.istio_webhook.dir=${dockerPushRoot}
global.helm_broker.version=${appVersion}
global.helm_broker.dir=${dockerPushRoot}
global.remote_environment_broker.version=${appVersion}
global.remote_environment_broker.dir=${dockerPushRoot}
global.metadata_service.version=${appVersion}
global.metadata_service.dir=${dockerPushRoot}
global.gateway.version=${appVersion}
global.gateway.dir=${dockerPushRoot}
global.installer.version=${appVersion}
global.installer.dir=${dockerPushRoot}
global.connector_service.version=${appVersion}
global.connector_service.dir=${dockerPushRoot}
global.ui_api_layer.version=${appVersion}
global.ui_api_layer.dir=${dockerPushRoot}
global.event_bus.version=${appVersion}
global.event_bus.dir=${dockerPushRoot}
global.event_service.version=${appVersion}
global.event_service.dir=${dockerPushRoot}
global.alpine_net.version=${appVersion}
global.alpine_net.dir=${dockerPushRoot}
global.watch_pods.version=${appVersion}
global.watch_pods.dir=${dockerPushRoot}
global.stability_checker.version=${appVersion}
global.stability_checker.dir=${dockerPushRoot}
global.etcd_backup.version=${appVersion}
global.etcd_backup.dir=${dockerPushRoot}
global.etcd_tls_setup.version=${appVersion}
global.etcd_tls_setup.dir=${dockerPushRoot}
global.test_logging_monitoring.version=${appVersion}
global.test_logging_monitoring.dir=${dockerPushRoot}
global.acceptance_tests.version=${appVersion}
global.acceptance_tests.dir=${dockerPushRoot}
global.ui_api_layer_acceptance_tests.version=${appVersion}
global.ui_api_layer_acceptance_tests.dir=${dockerPushRoot}
global.gateway_tests.version=${appVersion}
global.gateway_tests.dir=${dockerPushRoot}
global.test_environments.version=${appVersion}
global.test_environments.dir=${dockerPushRoot}
global.kubeless_test_client.version=${appVersion}
global.kubeless_test_client.dir=${dockerPushRoot}
global.api_controller_acceptance_tests.version=${appVersion}
global.api_controller_acceptance_tests.dir=${dockerPushRoot}
global.connector_service_tests.version=${appVersion}
global.connector_service_tests.dir=${dockerPushRoot}
global.metadata_service_tests.version=${appVersion}
global.metadata_service_tests.dir=${dockerPushRoot}
global.event_bus_tests.version=${appVersion}
global.event_bus_tests.dir=${dockerPushRoot}
global.test_logging.version=${appVersion}
global.test_logging.dir=${dockerPushRoot}
"""

    return "$overrides"
}
