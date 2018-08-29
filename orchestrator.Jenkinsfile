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
        - all component versions

*/
def label = "kyma-${UUID.randomUUID().toString()}"
appVersion = "0.3." + env.BUILD_NUMBER

/*
    Projects that are built when changed, consisting of pairs [project path, produced docker image]. Projects producing multiple Docker images only need to provide one of them.
    Projects that do NOT produce docker images need to have a null value and will not be passed to the integration job, as there is nothing to deploy.

    IMPORTANT NOTE: Projects trigger jobs and therefore are expected to have a job defined with the same name.
*/
projects = [
    "docs": "kyma-docs",
    "components/api-controller": "api-controller",
    "components/binding-usage-controller": "binding-usage-controller",
    "components/configurations-generator": "configurations-generator",
    "components/environments": "environments",
    "components/istio-webhook": "istio-webhook",
    "components/helm-broker": "helm-broker",
    "components/remote-environment-broker": "remote-environment-broker",
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
                                                string(name:'APP_VERSION', value: "$appVersion")
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

// gather all component versions
stage('collect versions') {
    versions = [:]
    
    changedProjects = jobs.keySet()
    for (int i = 0; i < changedProjects.size(); i++) {
        // only projects that have an associated docker image have a version to deploy
        if (projects["${changedProjects[i]}"] != null) {
            versions["${changedProjects[i]}"] = env.BRANCH_NAME == "master" ? appVersion : env.BRANCH_NAME
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
    try {
        def img = projects[project]
        def json = "https://eu.gcr.io/v2/kyma-project/${img}/manifests/latest".toURL().getText(requestProperties: [Accept: 'application/vnd.docker.distribution.manifest.v1+prettyjws'])
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
"""global:
  kyma:
    versions:
      docs: ${versions['docs']}
      api-controller: ${versions['components/api-controller']}
      binding-usage-controller: ${versions['components/binding-usage-controller']}
      configurations-generator: ${versions['components/configurations-generator']}
      environments: ${versions['components/environments']}
      istio-webhook: ${versions['components/istio-webhook']}
      helm-broker: ${versions['components/helm-broker']}
      remote-environment-broker: ${versions['components/remote-environment-broker']}
      metadata-service: ${versions['components/metadata-service']}
      gateway: ${versions['components/gateway']}
      installer: ${versions['components/installer']}
      connector-service: ${versions['components/connector-service']}
      ui-api-layer: ${versions['components/ui-api-layer']}
      event-bus: ${versions['components/event-bus']}
      alpine-net: ${versions['tools/alpine-net']}
      watch-pods: ${versions['tools/watch-pods']}
      stability-checker: ${versions['tools/stability-checker']}
      etcd-backup: ${versions['tools/etcd-backup']}
      test-logging-monitoring: ${versions['tests/test-logging-monitoring']}
      acceptance-tests: ${versions['tests/acceptance']}
      ui-api-layer-acceptance-tests: ${versions['tests/ui-api-layer-acceptance-tests']}
      gateway-tests: ${versions['tests/gateway-tests']}
      test-environments: ${versions['tests/test-environments']}
      kubeless-test-client: ${versions['tests/kubeless-test-client']}
      api-controller-acceptance-tests: ${versions['tests/api-controller-acceptance-tests']}
      connector-service-tests: ${versions['tests/connector-service-tests']}
      metadata-service-tests: ${versions['tests/metadata-service-tests']}
      event-bus-tests: ${versions['tests/event-bus']}

"""

    return "$overrides"
}
