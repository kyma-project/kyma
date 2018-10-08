#!/usr/bin/env sh

echo "Configuring git..."
git config --global user.email "kyma.bot@sap.com"
git config --global user.name "Kyma Bot"
git config --global core.sshCommand 'ssh -i '$SSH_FILE''
