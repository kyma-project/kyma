#!/usr/bin/env bash

readonly REPOS_DIR="/var/www/git"

mkdir "$REPOS_DIR"

git config --global user.email "gitserver@kyma-project.io"
git config --global user.name "Git Server"

for d in /tmp/repos/*/ ; do
  cd "$d" || exit 1
  git init && git add --all && git commit -m"initial commit"
  cd "$REPOS_DIR" || exit 1
  git clone --bare "$d"
done

chown -Rfv www-data:www-data /var/www/git
