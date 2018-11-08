def label = "kyma-${UUID.randomUUID().toString()}"
def registry = 'eu.gcr.io/kyma-project'
def acsImageName = 'acs-installer:0.0.4'

echo """********************************
Job started with the following parameters:
GIT_BRANCH=${params.GIT_BRANCH}
KYMA_INSTALLER_PUSH_DIR=${params.KYMA_INSTALLER_PUSH_DIR}
KYMA_INSTALLER_VERSION=${params.KYMA_INSTALLER_VERSION}
********************************
"""

podTemplate(label: label) {
    node(label) {
        timestamps {
            timeout(time:20, unit:"MINUTES") {

                currentBuild.displayName = "#${BUILD_NUMBER}, ${params.KYMA_INSTALLER_VERSION}"

                stage("Checkout") {
                    checkout scm
                }

                stage("Generate artifacts") {
                    def dockerEnv = "-e KYMA_INSTALLER_PUSH_DIR -e KYMA_INSTALLER_VERSION -e ARTIFACTS_DIR=/kyma"
                    def dockerOpts = "--rm --volume ${WORKSPACE}:/kyma"
                    def dockerEntry = "--entrypoint /kyma/installation/scripts/release-generate-kyma-installer-artifacts.sh"
                    
                    sh "docker run $dockerOpts $dockerEnv $dockerEntry $registry/$acsImageName"
                }

                archiveArtifacts artifacts: "kyma-config-cluster.yaml, kyma-config-local.yaml"
            }
        }
    }
}