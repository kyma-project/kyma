#!/bin/bash

ROOT_DIR="./docs/"
PUBLIC_DIR="$ROOT_DIR/public"
SCRIPT_NAME="move_to_public.sh"
ASSETS_DIR="$ROOT_DIR/assets"

mkdir -p "$PUBLIC_DIR"

# Find files that should be moved
find "$ROOT_DIR" -type f \
  ! -path "$ROOT_DIR/.vitepress/*" \
  ! -path "$ROOT_DIR/node_modules/*" \
  ! -path "$ROOT_DIR/.git/*" \
  ! -path "$ROOT_DIR/.github/*" \
  ! -path "$PUBLIC_DIR/*" \
  ! -name "*.md" \
  ! -name "*.png" \
  ! -name "*.jpg" \
  ! -name "*.jpeg" \
  ! -name "*.svg" \
  ! -name "*.webp" \
  ! -name "*.gif" \
  ! -name "*.ts" \
  ! -name "*.js" \
  ! -name "*.vue" \
  ! -name "*.css" \
  ! -name "*.scss" \
  ! -name "*.html" \
  ! -name "config.*" \
  ! -name ".DS_Store" \
  ! -name "$SCRIPT_NAME" \
  | while read -r file; do
    # Skip files in root directory unless they are inside /assets/
    if [[ "$(dirname "$file")" == "$ROOT_DIR" && "$file" != "$ASSETS_DIR"* ]]; then
      continue
    fi

    rel_path="${file#$ROOT_DIR}"
    target_path="$PUBLIC_DIR/$rel_path"

    mkdir -p "$(dirname "$target_path")"
    echo "Moving: $file → $target_path"
    cp "$file" "$target_path"
done

# Move entire /assets/ folder to public if it exists
if [ -d "$ASSETS_DIR" ]; then
  echo "Moving entire /assets/ folder to public/"
  cp -r "$ASSETS_DIR" "$PUBLIC_DIR"
fi
