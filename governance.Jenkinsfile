#!/usr/bin/env groovy
def label = "kyma-${UUID.randomUUID().toString()}"
def isMaster = params.GIT_BRANCH == 'master'
def repositoryName = "kyma"

echo """
********************************
Job started with the following parameters:
GIT_REVISION=${params.GIT_REVISION}
GIT_BRANCH=${params.GIT_BRANCH}
APP_VERSION=${params.APP_VERSION}
TRIGGER_FULL_VALIDATION=${params.TRIGGER_FULL_VALIDATION}
********************************
"""

podTemplate(label: label) {
    node(label) {
        try {
            timestamps {
                timeout(time:20, unit:"MINUTES") {
                    ansiColor('xterm') {
                        stage("setup") {
                            checkout scm
                        }


                        stage("validate internal links") {
                            validateLinks('--ignore-external', repositoryName)
                        }

                        if (isMaster) {
                            stage("validate external links") {
                                validateLinks('--ignore-internal', repositoryName)
                            }
                        } else {
                            stage("validate external links in changed markdown files") {
                                def changes = changedMarkdownFiles(repositoryName).join(" ")
                                validateLinks("--ignore-internal ${changes}", repositoryName)
                            }
                        }
                    }
                }
            }
        } catch (ex) {
            echo "Got exception: ${ex}"
            currentBuild.result = "FAILURE"
            def body = "${currentBuild.currentResult} ${env.JOB_NAME}${env.BUILD_DISPLAY_NAME}: on branch: ${params.GIT_BRANCH}. See details: ${env.BUILD_URL}"
            emailext body: body, recipientProviders: [[$class: 'DevelopersRecipientProvider'], [$class: 'CulpritsRecipientProvider'], [$class: 'RequesterRecipientProvider']], subject: "${currentBuild.currentResult}: Job '${env.JOB_NAME} [${env.BUILD_NUMBER}]'"

            if(isMaster) {
                sendSlackNotification(":hankey-fire:Governance validation failed on `${repositoryName}` repository!\nSee details: ${env.BUILD_URL}console")
            }
        }
    }
}

def validateLinks(args, repositoryName) {
    workDir = pwd()
    sh "docker run --rm --dns=8.8.8.8 --dns=8.8.4.4 -v ${workDir}:/${repositoryName}:ro magicmatatjahu/milv:0.0.6 --base-path=/${repositoryName} ${args}"
}

def sendSlackNotification(text) {
    def channel = "#kyma-readme-reviews"
    echo "Sending notification on Slack to channel: ${channel}"

    withCredentials([string(credentialsId: 'kyma-slack-token', variable: 'token')]) {
        sh """
            curl -H 'Content-type: application/json' \
                --data '{"text": "${text}", "channel": "${channel}"}' \
                https://sap-cx.slack.com/services/hooks/jenkins-ci?token=${token}
        """
    }
}

/* -------- Helper Functions -------- */

/**
 * Provides a list with the .md files that have changed
 */
String[] changedMarkdownFiles(String repositoryName) {
    res = []
    echo "Looking for changes in the markdown files"

    // get all changes
    allChanges = changeset().split("\n")

    // if no changes
    if (allChanges.size() == 0) {
        echo "No changes found or could not be fetched"
        return res
    }

    // add ${repositoryName} suffix to markdown path
    for (int i=0; i < allChanges.size(); i++) {
        res.add("./${repositoryName}/${allChanges[i]}")
    }
    return res
}

/**
 * Gets the changes on the Project based on the branch
 */
@NonCPS
String changeset() {
    prPrefix = 'PR-';
    branch = params.GIT_BRANCH.substring(prPrefix.size())

    // get changeset comparing branch with master
    echo "Fetching changes between remotes/origin/${branch}/head and remotes/origin/master."
    return sh (script: "git --no-pager diff --name-only remotes/origin/master...remotes/origin/${branch}/head | grep '.md' || echo ''", returnStdout: true)
}
