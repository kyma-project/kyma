#!/usr/bin/env groovy
def label = "kyma-${UUID.randomUUID().toString()}"

echo """
********************************
Job started with the following parameters:
GIT_REVISION=${params.GIT_REVISION}
GIT_BRANCH=${params.GIT_BRANCH}
APP_VERSION=${params.APP_VERSION}
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

                        stage("validate external links") {
                            validateLinks('find . -name "*.md" | grep -v "vendor"')
                        }
                    }
                }
            }
        } catch (ex) {
            echo "Got exception: ${ex}"
            currentBuild.result = "FAILURE"
            def body = "${currentBuild.currentResult} ${env.JOB_NAME}${env.BUILD_DISPLAY_NAME}: on branch: ${params.GIT_BRANCH}. See details: ${env.BUILD_URL}"
            emailext body: body, recipientProviders: [[$class: 'DevelopersRecipientProvider'], [$class: 'CulpritsRecipientProvider'], [$class: 'RequesterRecipientProvider']], subject: "${currentBuild.currentResult}: Job '${env.JOB_NAME} [${env.BUILD_NUMBER}]'"
        }
    }
}

def validateLinks(command) {
    workDir = pwd()
    whiteList = "orders.com,github.com,localhost,alert.victorops.com,hooks.slack.com,named-lynx-etcd,gateway.CLUSTER,connector-service.CLUSTER,svc.cluster.local,10.0.0.54,10.0.0.55,192.168.64.44,minio,dex-web,dex-service,http-db-service,custom.bundles-repository,grafana,testjs.default,abc.com,sampleapis.com,regularsampleapis.com,httpbin.org,console.kyma.local,kyma.cx,ec.com,github.io,dummy.url,kyma.local,jaegertracing.io,semver.org,example.com"
    sh "docker run --rm -v $workDir:/mnt:ro dkhamsing/awesome_bot --allow-dupe --allow-redirect --skip-save-results --allow-ssl --white-list $whiteList `$command`"
}