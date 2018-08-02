#!/usr/bin/env groovy
/*

Monorepo root orchestrator: This Jenkinsfile runs the Jenkinsfiles of all subprojects based on the changes made and triggers kyma integration.
    - checks for changes since last successful build on master and compares to master if on a PR.
    - for every changed project, triggers related job async as configured in the seedjob.
    - for every changed additional project, triggers the kyma integration job.
    - passes info of:
        - revision
        - branch
        - current app version

*/
def label = "kyma-${UUID.randomUUID().toString()}"
appVersion = "0.3." + env.BUILD_NUMBER

/*
    Projects that are built when changed.
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
    "components/application-connector",
    "components/gateway",
    "components/installer",
    "components/connector-service",
    "components/ui-api-layer",
    "components/event-bus",
    "tools/alpine-net",
    "tools/watch-pods",
    "tools/stability-checker",
    "tools/etcd-backup",
    "tests/test-logging-monitoring",
    "tests/acceptance",
    "tests/ui-api-layer-acceptance-tests",
    "tests/application-connector-tests",
    "tests/gateway-tests",
    "tests/test-environments",
    "tests/kubeless-test-client",
    "tests/api-controller-acceptance-tests",
    "tests/connector-service-tests",
    "tests/event-bus",
    "governance" 
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
                        buildableProjects = changes.intersect(projects) // only projects that have build jobs
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

// trigger jobs for projects that have changes, in parallel
stage('build projects') {
    parallel jobs
}

// trigger Kyma integration when changes are made to installation charts/code or resources
if (runIntegration) {
    stage('launch Kyma integration') {
        build job: 'kyma/integration',
            wait: false,
            parameters: [
                string(name:'GIT_REVISION', value: "$commitID"),
                string(name:'GIT_BRANCH', value: "${env.BRANCH_NAME}"),
                string(name:'APP_VERSION', value: "$appVersion")
            ]
    }
}



/* -------- Helper Functions -------- */

/**
 * Provides a list with the projects that have changes within the given projects list.
 * If no changes found, all projects will be returned.
 */
String[] changedProjects() {
    res = []
    def allProjects = projects + additionalProjects
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
            if (projects[i] == "governance" && allChanges[j].endsWith(".md") && !res.contains(projects[i])) {
                res.add(projects[i])
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
