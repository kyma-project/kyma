#!/bin/bash

# === CONFIGURAZIONE ===
ORG="kyma-project"
#TOKEN="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"
PER_PAGE=100
PAGE=1

# === HEADERS ===
#HEADER="Authorization: token $TOKEN"

# === CREAZIONE CARTELLA ===
mkdir -p "$ORG"
cd "$ORG" || exit

echo "Recupero repository da GitHub per l'organizzazione: $ORG..."

while : ; do
    RESPONSE=$(curl -s "https://api.github.com/orgs/$ORG/repos?per_page=$PER_PAGE&page=$PAGE")

    # Estrai gli URL di clonazione usando jq
    REPO_URLS=$(echo "$RESPONSE" | jq -r '.[].clone_url')

    # Se non ci sono URL, termina
    if [ -z "$REPO_URLS" ]; then
        echo "Fine dei repository."
        break
    fi

    # Clona ogni repository
    while IFS= read -r REPO; do
        echo "Clonando $REPO..."
        git clone "$REPO"
    done <<< "$REPO_URLS"

    PAGE=$((PAGE + 1))
done

echo "✅ Clonazione completata."

# === ESTRAZIONE CARTELLE docs/user ===
echo "🔍 Estrazione cartelle docs/user dai repository..."

mkdir -p ../modules

for dir in */ ; do
    REPO_NAME="${dir%/}"
    SOURCE_PATH="$REPO_NAME/docs/user"
    TARGET_PATH="../modules/$REPO_NAME/docs/user"
    SOURCE_PATH_ASSETS="$REPO_NAME/docs/assets"
    TARGET_PATH_ASSETS="../modules/$REPO_NAME/docs/assets"

    if [ -d "$SOURCE_PATH" ]; then
        echo "📁 Trovata in $REPO_NAME, copio in modules/$REPO_NAME/docs/user"
        mkdir -p "$TARGET_PATH"
        cp -r "$SOURCE_PATH/" "$TARGET_PATH/"
        if [ -d "$SOURCE_PATH_ASSETS" ]; then
            echo "↳📁 Trovata in $REPO_NAME, copio in modules/$REPO_NAME/docs/assets"
            mkdir -p "$TARGET_PATH_ASSETS"
            cp -r "$SOURCE_PATH_ASSETS/" "$TARGET_PATH_ASSETS/"
        else
            echo "↳🚫 Nessuna cartella docs/assets in $REPO_NAME"
        fi
    else
        echo "🚫 Nessuna cartella docs/users in $REPO_NAME"
    fi
done

echo "✅ Operazione completata."
