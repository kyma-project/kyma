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
def registry = 'eu.gcr.io/kyma-project'
def acsImageName = 'acs-installer:0.0.4'
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
    "components/istio-kyma-patch": "istio-kyma-patch",
    "components/helm-broker": "helm-broker",
    "components/application-broker": "application-broker",
    "components/application-operator": "application-operator",
    "components/metadata-service": "metadata-service",
    "components/installer": "installer",
    "components/connector-service": "connector-service",
    "components/connection-token-handler": "connection-token-handler",
    "components/ui-api-layer": "ui-api-layer",
    "components/event-bus": "event-bus-publish",
    "components/event-service": "event-service",
    "components/proxy-service": "proxy-service",
    "components/k8s-dashboard-proxy": "k8s-dashboard-proxy",
    "tools/alpine-net": "alpine-net",
    "tools/watch-pods": "watch-pods",
    "tools/stability-checker": "stability-checker",
    "tools/etcd-backup": "etcd-backup",
    "tools/etcd-tls-setup": "etcd-tls-setup",
    "tools/gcp-broker-provider": "gcp-broker-provider",
    "tools/static-users-generator": "static-users-generator",
    "tools/ark-plugins": "ark-plugins",
    "tests/test-logging-monitoring": "test-logging-monitoring",
    "tests/logging": "test-logging",
    "tests/acceptance": "acceptance-tests",
    "tests/ui-api-layer-acceptance-tests": "ui-api-layer-acceptance-tests",
    "tests/gateway-tests": "gateway-acceptance-tests",
    "tests/test-environments": "test-environments",
    "tests/kubeless-integration": "kubeless-integration-tests",
    "tests/kubeless": "kubeless-tests",
    "tests/api-controller-acceptance-tests": "api-controller-acceptance-tests",
    "tests/connector-service-tests": "connector-service-tests",
    "tests/metadata-service-tests": "metadata-service-tests",
    "tests/application-operator-tests": "application-operator-tests",
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
    NOTE: This is set automatically based on the changes detected.
*/
runIntegration = false

properties([
    buildDiscarder(logRotator(numToKeepStr: '10')),
    disableConcurrentBuilds()
])

try {
    podTemplate(label: label) {
        node(label) {
            timestamps {
                ansiColor('xterm') {
                    stage("setup") {
                        checkout scm
                        // use HEAD of branch as revision, Jenkins does a merge to master commit before starting this script, which will not be available on the jobs triggered below
                        commitID = sh (script: "git rev-parse origin/${env.BRANCH_NAME}", returnStdout: true).trim()
                        configureBuilds(commitID)
                        changes = changedProjects()

                        if (isMaster) {
                            // integration runs on any change on master
                            runIntegration = changes.size() > 0
                        } else {
                            // integration only runs on changes to installation resources on PRs
                            runIntegration = changes.intersect(additionalProjects).size() > 0
                        }

                        if (changes.size() == 1 && (changes[0] == "governance" || changes[0] == "docs")){
                            runIntegration = false
                        }

                        if (changes.size() == 2 && changes.contains("governance") && changes.contains("docs")) {
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
                    string(name:'APP_VERSION', value: "$appVersion")
                ]
        }
    }
}  catch (ex) {
    echo "Got exception: ${ex}"
    currentBuild.result = "FAILURE"
    def body = "${currentBuild.currentResult} ${env.JOB_NAME}${env.BUILD_DISPLAY_NAME}: on branch: ${env.BRANCH_NAME}. See details: ${env.BUILD_URL}"
    emailext body: body, recipientProviders: [[$class: 'DevelopersRecipientProvider'], [$class: 'CulpritsRecipientProvider'], [$class: 'RequesterRecipientProvider']], subject: "${currentBuild.currentResult}: Job '${env.JOB_NAME} [${env.BUILD_NUMBER}]'"
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
            if (env.BRANCH_NAME != 'master' && allProjects[i] == "governance" && allChanges[j].endsWith(".md") && !res.contains(allProjects[i])) {
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
    if (env.BRANCH_NAME != 'master') {
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
