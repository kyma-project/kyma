#!/bin/bash

ORG=kyma-project
TARGET_DIR=../docs/externalContent
mkdir -p "$ORG"
mkdir -p "$TARGET_DIR"
cd "$ORG" || exit

echo '🔁 Cloning selected repositories...'
REPOS=(
  btp-manager
  istio
  serverless
  telemetry-manager
  eventing-manager
  api-gateway
  nats-manager
  application-connector-manager
  keda-manager
  cloud-manager
  docker-registry
  busola
  cli
)

for repo in "${REPOS[@]}"; do
  echo "📥 Cloning https://github.com/$ORG/$repo.git"
  git clone https://github.com/$ORG/$repo.git
done

cd ..
echo '📂 Copying docs/user and docs/assets folders...'
for repo in "${REPOS[@]}"; do
  SOURCE_USER="$ORG/$repo/docs/user"
  TARGET_USER="$TARGET_DIR/$repo/docs/user"
  SOURCE_ASSETS="$ORG/$repo/docs/assets"
  TARGET_ASSETS="$TARGET_DIR/$repo/docs/assets"

  if [ -d "$SOURCE_USER" ]; then
    echo "📁 Copying $SOURCE_USER to $TARGET_USER"
    mkdir -p "$TARGET_USER"
    cp -r "$SOURCE_USER/" "$TARGET_USER/"
  else
    echo "🚫 No docs/user folder in $repo"
  fi

  if [ -d "$SOURCE_ASSETS" ]; then
    echo "📁 Copying $SOURCE_ASSETS to $TARGET_ASSETS"
    mkdir -p "$TARGET_ASSETS"
    cp -r "$SOURCE_ASSETS/" "$TARGET_ASSETS/"
  else
    echo "🚫 No docs/assets folder in $repo"
  fi
done

echo '🧹 Cleanup: remove all cloned repositories...'
rm -rf "$ORG"

echo '✅ Operation completed.'