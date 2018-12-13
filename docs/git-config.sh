#!/usr/bin/env bash
set -e
set -o pipefail

SSH_FILE=
while test $# -gt 0; do
    case "$1" in
        --ssh-file | -s)
            shift
            SSH_FILE=$1
            shift
            ;;
        *)
            echo "$1 is not a recognized flag!"
            exit 1;
            ;;
    esac
done
readonly SSH_FILE

cd /home/prow/go/src/github.com/kyma-project/website
# configure ssh
mkdir "${HOME}/.ssh/"
touch "${HOME}/.ssh/known_hosts"
ssh-keyscan -H github.com >> "${HOME}/.ssh/known_hosts"
chmod 400 "${SSH_FILE}"
eval "$(ssh-agent -s)"
ssh-add "${SSH_FILE}"
ssh-add -l

# configure git
git config --global user.email "${BOT_GITHUB_EMAIL}"
git config --global user.name "${BOT_GITHUB_NAME}"
git config --global core.sshCommand 'ssh -i '${SSH_FILE}''
echo $(pwd)
# git remote add origin git@github.com:kyma-project/website.git
