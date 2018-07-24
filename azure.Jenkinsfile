def label = sanitizeLabel(env.JOB_NAME, env.BUILD_NUMBER)
def isAzurePodCreated = false

properties([
    buildDiscarder(logRotator(daysToKeepStr: '14', numToKeepStr: '30')),
    disableConcurrentBuilds()
])

podTemplate(label: label) {
    node(label) {
        try {
            timestamps {
                timeout(time:40, unit:"MINUTES") {
                    ansiColor('xterm') {
                        stage("cleanup") {
                            cleanup()
                        }

                        stage("checkout kyma scm") {
                            dir("kyma") {
                                checkout scm
                            }
                        }

                        stage("build image") {
                            dir("kyma") {
                                sh "installation/cmd/ci.build.sh"
                            }
                        }

                        env.IMAGE_TAG = UUID.randomUUID().toString()

                        stage("push imige to ACR") {
                            withCredentials([usernamePassword(credentialsId: 'azure-kyma-spn', passwordVariable: 'ARM_CLIENT_SECRET', usernameVariable: 'ARM_CLIENT_ID'),
                                            string(credentialsId: 'ci-azure-cluster-kubeconfig', variable: 'KUBECONFIG_JSON'),
                                            string(credentialsId: 'ci-azure-cluster-acr-name', variable: 'ACR_NAME')]) {
                                dir("kyma") {
                                    sh "installation/cmd/azure-ci.run.sh pushDockerImageToAcr"
                                }
                            }
                        }

                        stage("run image on Azure k8s") {
                            isAzurePodCreated = true
                            withCredentials([usernamePassword(credentialsId: 'azure-kyma-spn', passwordVariable: 'ARM_CLIENT_SECRET', usernameVariable: 'ARM_CLIENT_ID'),
                                            string(credentialsId: 'ci-azure-cluster-kubeconfig', variable: 'KUBECONFIG_JSON'),
                                            string(credentialsId: 'ci-azure-cluster-acr-name', variable: 'ACR_NAME')]) {
                                def dockerEnv = "-e ACR_NAME -e KUBECONFIG_JSON -e IMAGE_TAG"
                                def dockerOpts = "-w='/kyma' --rm --volume ${WORKSPACE}/kyma:/kyma"
                                def dockerEntry = "--entrypoint /kyma/installation/cmd/azure-ci.run.sh"
                                def dockerEntryArg = "createPod"
                                sh "docker run $dockerOpts $dockerEnv $dockerEntry kyma-on-minikube:latest $dockerEntryArg"
                            }
                        }

                        if (env.BRANCH_NAME == "master" && currentBuild.getPreviousBuild() && currentBuild.getPreviousBuild().getResult().toString() != "SUCCESS") {
                            echo "\033[32m Sending RECOVERY message on Slack \033[0m\n"
                            def recipients = emailextrecipients([[$class: 'DevelopersRecipientProvider'], [$class: 'RequesterRecipientProvider']])
                            sendSlackNotification(":white_check_mark:Kyma heroes who made [Azure] Kyma LOCAL great again: ${recipients}!\nSee details: ${env.BUILD_URL}console")
                        }
                    }
                }
            }
        } catch (ex) {
            echo "Got exception: ${ex}"
            currentBuild.result = "FAILURE"
            def body = """${currentBuild.currentResult} ${env.JOB_NAME}${env.BUILD_DISPLAY_NAME}: on branch: ${env.BRANCH_NAME}. See details: ${env.BUILD_URL}"""
            emailext body: body, recipientProviders: [[$class: 'DevelopersRecipientProvider'], [$class: 'CulpritsRecipientProvider'], [$class: 'RequesterRecipientProvider']], subject: "[AZURE] Kyma local: ${currentBuild.currentResult}: Job '${env.JOB_NAME} [${env.BUILD_NUMBER}]'"
            if (env.BRANCH_NAME == "master") {
                def culprits = emailextrecipients([[$class: 'CulpritsRecipientProvider'], [$class: 'RequesterRecipientProvider']])
                sendSlackNotification(":x:[Azure] Kyma LOCAL BUILD FAILED! Committers since last success build: ${culprits}\nSee details: ${env.BUILD_URL}console")
            }
        }
        finally {
            if (isAzurePodCreated) {
                echo "--- Delete minikube pod in azure cluster ---"
                withCredentials([string(credentialsId: 'ci-azure-cluster-kubeconfig', variable: 'KUBECONFIG_JSON')]) {
                    def dockerEnv = "-e KUBECONFIG_JSON -e IMAGE_TAG"
                    def dockerOpts = "-w='/kyma' --rm --volume ${WORKSPACE}/kyma:/kyma"
                    def dockerEntry = "--entrypoint /kyma/installation/cmd/azure-ci.run.sh"
                    def dockerEntryArg = "deletePod"
                    sh "docker run $dockerOpts $dockerEnv $dockerEntry kyma-on-minikube:latest $dockerEntryArg"
                }
            }
        }
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

def sanitizeLabel(label, number) {
    def labelSanitized = label.replaceAll(/[^-_.A-Za-z0-9]/, '_').take(62 - number.toString().size())
    "a${labelSanitized}${number}"
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
