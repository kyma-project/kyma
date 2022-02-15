#!/usr/bin/env sh

readonly REPOS_DIR="/var/www/git"

mkdir "$REPOS_DIR"

for d in /tmp/repos/*/ ; do
  cd "$d" || exit 1
  git init && git add --all && git commit -m"initial commit"
  cd "$REPOS_DIR" || exit 1
  git clone --bare "$d"
done

chown -Rfv apache:apache /var/www/git
