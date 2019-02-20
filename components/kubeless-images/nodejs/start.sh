#!/bin/sh

echo "Lambda Code"
awk '{gsub(/\\/,"\\\\\\\\", $0); print $0}' /kubeless/handler.js | awk '{printf "%s\\r\\n", $0}' | rev | cut -c 5- | rev

echo "\n"
if [ -e  /kubeless/package.json ]; then
  echo "Lambda Code dependencies"
  awk '{gsub(/\\/,"\\\\\\\\", $0); print $0}' /kubeless/package.json | awk '{printf "%s\\r\\n", $0}'| rev | cut -c 5- | rev
fi

node kubeless.js