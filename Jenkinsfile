#!/usr/bin/env groovy
def label = "kyma-${UUID.randomUUID().toString()}"
def commit_id=''
def isMaster = params.GIT_BRANCH == 'master'

echo """
********************************
Job started with the following parameters:
DOCKER_REGISTRY=${env.DOCKER_REGISTRY}
DOCKER_CREDENTIALS=${env.DOCKER_CREDENTIALS}
GIT_REVISION=${params.GIT_REVISION}
GIT_BRANCH=${params.GIT_BRANCH}
APP_VERSION=${params.APP_VERSION}
********************************
"""

podTemplate(label: label) {
    node(label) {
        try {
            timestamps {
                ansiColor('xterm') {
                    timeout(time:40, unit:"MINUTES") {
                            stage("cleanup") {
                                cleanup()
                            }

                            stage("checkout kyma") {
                                dir("kyma") {
                                    checkout scm
                                }
                            }

                            stage("build image") {
                                dir("kyma") {
                                    sh "installation/cmd/ci.build.sh"
                                }
                            }

                            stage("configure and test container") {
                                dir("kyma") {
                                    sh "installation/cmd/ci.run.sh --non-interactive --exit-on-test-fail"
                                }
                            }

                        if (isMaster && currentBuild.getPreviousBuild() && currentBuild.getPreviousBuild().getResult().toString() != "SUCCESS") {
                            echo "\033[32m Sending RECOVERY message on Slack \033[0m\n"
                            def recipients = emailextrecipients([[$class: 'DevelopersRecipientProvider'], [$class: 'RequesterRecipientProvider']])
                            sendSlackNotification(":white_check_mark:Kyma heroes who made Kyma LOCAL great again: ${recipients}!\nSee details: ${env.BUILD_URL}console")
                        }
                    }

                    if (isMaster) {
                        stage("save revision") {
                            dir("kyma") {
                                sh "git rev-parse --short HEAD > .git/commit-id"
                                commit_id = readFile('.git/commit-id')
                                commit_id = commit_id.trim()
                            }
                        }
                    }
                }
            }
        } catch (ex) {
            echo "Got exception: ${ex}"
            currentBuild.result = "FAILURE"
            def body = """${currentBuild.currentResult} ${env.JOB_NAME}${env.BUILD_DISPLAY_NAME}: on branch: ${params.GIT_BRANCH}. See details: ${env.BUILD_URL}"""
            emailext body: body, recipientProviders: [[$class: 'DevelopersRecipientProvider'], [$class: 'CulpritsRecipientProvider'], [$class: 'RequesterRecipientProvider']], subject: "Kyma: ${currentBuild.currentResult}: Job '${env.JOB_NAME} [${env.BUILD_NUMBER}]'"
            if (isMaster) {
                def culprits = emailextrecipients([[$class: 'CulpritsRecipientProvider'], [$class: 'RequesterRecipientProvider']])
                sendSlackNotification(":x:Kyma LOCAL BUILD FAILED! Committers since last success build: ${culprits}\nSee details: ${env.BUILD_URL}console")
            }
        }
    }
}

if(isMaster && currentBuild.currentResult == "SUCCESS") {
    stage("trigger remote cluster"){
        build job: 'azure/master', parameters: [
            string(name:'REVISION', value: "${commit_id}")],
            wait: false
    }
}

def sendSlackNotification(text) {
    def channel = "#c4core-kyma-team"
    echo "Sending notification on Slack to channel: ${channel}"

    withCredentials([string(credentialsId: 'kyma-slack-token', variable: 'token')]) {
        sh """
            curl -H 'Content-type: application/json' \
                --data '{"text": "${text}", "channel": "${channel}"}' \
                https://sap-cx.slack.com/services/hooks/jenkins-ci?token=${token}
        """
    }
}

def cleanup(target = '') {
    if (target) {
        echo "cleaning up ${target}"
        node(target) {
            sh 'find . -not -name "." -not -name ".." -maxdepth 1 -exec rm -rf {} \\; '
        }
    } else {
        echo "cleaning up current node"
        sh 'find . -not -name "." -not -name ".." -maxdepth 1 -exec rm -rf {} \\; '
    }
}