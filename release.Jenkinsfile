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
def registry = 'eu.gcr.io/kyma-project'
def acsImageName = 'acs-installer:0.0.4'

semVerRegex = /^([0-9]+\.[0-9]+\.[0-9]+)$/ // semVer format: 1.2.3
releaseBranchRegex = /^release\-([0-9]+\.[0-9]+)$/ // release branch format: release-1.5
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
    "components/apiserver-proxy",
    "components/binding-usage-controller",
    "components/configurations-generator",
    "components/environments",
    "components/istio-kyma-patch",
    "components/helm-broker",
    "components/application-broker",
    "components/application-operator",
    "components/metadata-service",
    "components/installer",
    "components/connector-service",
    "components/connection-token-handler",
    "components/ui-api-layer",
    "components/event-bus",
    "components/event-service",
    "components/proxy-service",
    "tools/alpine-net",
    "tools/watch-pods",
    "tools/stability-checker",
    "tools/etcd-backup",
    "tools/etcd-tls-setup",
    "tools/static-users-generator",
    "tests/test-logging-monitoring",
    "tests/logging",
    "tests/acceptance",
    "tests/ui-api-layer-acceptance-tests",
    "tests/gateway-tests",
    "tests/test-environments",
    "tests/kubeless-integration",
    "tests/kubeless",
    "tests/api-controller-acceptance-tests",
    "tests/connector-service-tests",
    "tests/metadata-service-tests",
    "tests/application-operator-tests",
    "tests/event-bus"
]

/*
    project jobs to run are stored here to be sent into the parallel block outside the node executor.
*/
jobs = [:]

try {
    podTemplate(label: label) {
        node(label) {
            timestamps {
                ansiColor('xterm') {
                    stage("Setup") {
                        checkout scm

                        // validate parameters
                        if (!isRelease && !params.RELEASE_VERSION.isEmpty()) {
                            error("Release version ${params.RELEASE_VERSION} does not follow semantic versioning.")
                        }
                        if (!params.RELEASE_BRANCH ==~ releaseBranchRegex) {
                            error("Release branch ${params.RELEASE_BRANCH} is not a valid branch. Provide a branch such as 'release-1.5'")
                        }
                    
                        commitID = sh (script: "git rev-parse origin/${params.RELEASE_BRANCH}", returnStdout: true).trim()
                        configureBuilds()
                    }

                    if(params.BUILD_COMPONENTS) {
                        stage('Collect projects') {
                            for (int i=0; i < projects.size(); i++) {
                                def index = i
                                jobs["${projects[index]}"] = { ->
                                        build job: "kyma/"+projects[index]+"-release",
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

                        // build components
                        stage('Build projects') {
                            parallel jobs
                        }
                    }

                    // build kyma-installer
                    stage('Build kyma-installer') {
                        build job: 'kyma/kyma-installer',
                            wait: true,
                            parameters: [
                                string(name:'GIT_BRANCH', value: "${params.RELEASE_BRANCH}"),
                                string(name:'PUSH_DIR', value: "$dockerPushRoot"),
                                string(name:'APP_VERSION', value: "$appVersion")
                            ]
                    }

                    // generate kyma-installer artifacts
                    def kymaInstallerArtifactsBuild = null
                    stage('Generate kyma-installer artifacts') {
                        kymaInstallerArtifactsBuild = build job: 'kyma/kyma-installer-artifacts',
                            wait: true,
                            parameters: [
                                string(name:'GIT_BRANCH', value: "${params.RELEASE_BRANCH}"),
                                string(name:'KYMA_INSTALLER_PUSH_DIR', value: "$dockerPushRoot"),
                                string(name:'KYMA_INSTALLER_VERSION', value: "$appVersion")
                            ]
                    }

                    stage('Copy kyma-installer artifacts') {
                        copyArtifacts projectName: 'kyma/kyma-installer-artifacts',
                            selector: specific("${kymaInstallerArtifactsBuild.number}"),
                            target: "kyma-installer-artifacts"
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

/* -------- Helper Functions -------- */

/** Configure the parameters for the components to build:
 * - release candidate: push root: "rc/" / image tag: short commit ID
 * - release: push root: "" / image tag: semantic version
 */
def configureBuilds() {
    if(isRelease) {
        echo ("Building Release ${params.RELEASE_VERSION}")
        dockerPushRoot = ""
        appVersion = params.RELEASE_VERSION
    } else {
        echo ("Building Release Candidate for ${params.RELEASE_BRANCH}")
        dockerPushRoot = "rc/"
        appVersion = "${(params.RELEASE_BRANCH =~ /([0-9]+\.[0-9]+)$/)[0][1]}-rc" // release branch number + '-rc' suffix (e.g. 1.0-rc)
    }   
}