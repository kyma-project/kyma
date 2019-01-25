#!/usr/bin/env groovy
import groovy.json.JsonSlurperClassic

def label = "kyma-${UUID.randomUUID().toString()}"
def registry = 'eu.gcr.io/kyma-project'
def acsImageName = 'acs-installer:0.0.4'
def changelogGeneratorPath = "tools/changelog-generator"

echo """
********************************
Job started with the following parameters:
RELEASE_VERSION=${env.RELEASE_VERSION}
RELEASE_BRANCH=${env.RELEASE_BRANCH}
ARTIFACTS_BUILD_NUMBER=${params.ARTIFACTS_BUILD_NUMBER}
********************************
"""

semVerRegex = /^([0-9]+\.[0-9]+\.[0-9]+)$/ // semVer format: 1.2.3
releaseBranchRegex = /^release\-([0-9]+\.[0-9]+)$/ // release branch format: release-1.5
isRelease = params.RELEASE_VERSION ==~ semVerRegex

commitID = ''
appVersion = ''

podTemplate(label: label) {
    node(label) {
        timestamps {
            ansiColor('xterm') {
                stage("Setup") {
                    checkout scm

                    if (!isRelease && !params.RELEASE_VERSION.isEmpty()) {
                        error("Release version ${params.RELEASE_VERSION} does not follow semantic versioning.")
                    }
                    if (!params.RELEASE_BRANCH ==~ releaseBranchRegex) {
                        error("Release branch ${params.RELEASE_BRANCH} is not a valid branch. Provide a branch such as 'release-1.5'")
                    }
                
                    commitID = sh (script: "git rev-parse origin/${params.RELEASE_BRANCH}", returnStdout: true).trim()

                    if(isRelease) {
                        echo ("Building Release ${params.RELEASE_VERSION}")
                        appVersion = params.RELEASE_VERSION
                    } else {
                        echo ("Building Release Candidate for ${params.RELEASE_BRANCH}")
                        appVersion = "${(params.RELEASE_BRANCH =~ /([0-9]+\.[0-9]+)$/)[0][1]}-rc" // release branch number + '-rc' suffix (e.g. 1.0-rc)
                    }
                }

                stage("Copy artifacts") {
                    copyArtifacts projectName: 'kyma/kyma-installer-artifacts', 
                        selector: specific("${params.ARTIFACTS_BUILD_NUMBER}"),
                        target: "kyma-installer-artifacts"
                }

                stage("Publish ${isRelease ? 'Release' : 'Prerelease'} ${appVersion}") {
                    
                    withCredentials(
                            [string(credentialsId: 'public-github-token', variable: 'token'),
                            sshUserPrivateKey(credentialsId: "bitbucket-rw", keyFileVariable: 'sshfile')
                    ]) {
                        // Build changelog generator
                        dir(changelogGeneratorPath) {
                            sh "docker build -t changelog-generator ."
                        }

                        // Setup path to repository config file for `lerna-changelog`
                        def configFile = "./tools/changelog-generator/package.json"

                        // Generate release changelog
                        changelogGenerator('/app/generate-release-changelog.sh --configure-git', ["NEW_RELEASE_TITLE=${appVersion}", "GITHUB_AUTH=${token}", "SSH_FILE=${sshfile}", "CONFIG_FILE=${configFile}"])

                        // Generate CHANGELOG.md
                        changelogGenerator('/app/generate-full-changelog.sh --configure-git', ["NEW_RELEASE_TITLE=${appVersion}", "GITHUB_AUTH=${token}", "SSH_FILE=${sshfile}", "CONFIG_FILE=${configFile}"])
                        sh "BRANCH=${params.RELEASE_BRANCH} NEW_RELEASE_TITLE=${appVersion} SSH_FILE=${sshfile} APP_PATH=./tools/changelog-generator/app ./tools/changelog-generator/app/push-full-changelog.sh --configure-git"
                        commitID = sh (script: "git rev-parse HEAD", returnStdout: true).trim()

                        def releaseChangelog = readFile "./.changelog/release-changelog.md"
                        def body = releaseChangelog.replaceAll("(\\r|\\n|\\r\\n)+", "\\\\n")

                        def data = new JsonSlurperClassic().parseText('{"tag_name": "","target_commitish": "","name": "","body": "","draft": false ,"prerelease": ""}')
                        data.tag_name = "${appVersion}"
                        data.target_commitish = "${commitID}"
                        data.name = "${appVersion}"
                        data.body = "${body}"
                        data.prerelease = isRelease ? 'false' : 'true'

                        sh "echo \"${data}\" > data.json"

                        echo "Creating a new release using GitHub API..."
                        def json = sh (script: "curl --data @data.json -H \"Authorization: token ${token}\" https://api.github.com/repos/kyma-project/kyma/releases", returnStdout: true)
                        echo "Response: ${json}"
                        def releaseID = getGithubReleaseID(json)

                        // upload artifacts
                        def kymaConfigLocal = "kyma-installer-artifacts/kyma-config-local.yaml"
                        def kymaConfigCluster = "kyma-installer-artifacts/kyma-config-cluster.yaml"

                        sh "curl --data-binary @${kymaConfigLocal} -H \"Authorization: token ${token}\" -H \"Content-Type: application/x-yaml\" https://uploads.github.com/repos/kyma-project/kyma/releases/${releaseID}/assets?name=kyma-config-local.yaml"
                        sh "curl --data-binary @${kymaConfigCluster} -H \"Authorization: token ${token}\" -H \"Content-Type: application/x-yaml\" https://uploads.github.com/repos/kyma-project/kyma/releases/${releaseID}/assets?name=kyma-config-cluster.yaml"
                    }
                }
            }
        }
    }
}

/**
 * Obtain the github release ID from its JSON data.
 * More info: https://developer.github.com/v3/repos/releases 
 */
@NonCPS
def getGithubReleaseID(releaseJson) {
    def slurper = new JsonSlurperClassic()
    return slurper.parseText(releaseJson).id
}

def changelogGenerator(command, envs = []) {
    def repositoryName = 'kyma'
    def image = 'changelog-generator'
    def envText = ''
    for (it in envs) {
        envText = "$envText --env $it"
    }
    workDir = pwd()

    def dockerRegistry = env.DOCKER_REGISTRY
    sh "docker run --rm -v $workDir:/$repositoryName -w /$repositoryName $envText $image sh $command"
}