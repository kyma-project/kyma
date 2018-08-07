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

                        if(isMaster || params.TRIGGER_FULL_VALIDATION) {
                            stage("validate external links") {
                                validateLinks('--ignore-internal', repositoryName)
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
    sh "docker run --rm -v ${workDir}:/${repositoryName}:ro magicmatatjahu/milv:0.0.3 --base-path=/${repositoryName} ${args}"
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