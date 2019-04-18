#!/bin/bash
echo "----------------------------------------------"
echo "Lambda Code"
awk '{gsub(/\\/,"\\\\\\\\", $0); printf "%s\\n", $0}' /kubeless/handler.js | rev | cut -c 3- | rev
echo
if [ -e  /kubeless/package.json ]; then
  echo "Lambda Code dependencies are"
  awk '{gsub(/\\/,"\\\\\\\\", $0); printf "%s\\n", $0}' /kubeless/package.json | rev | cut -c 3- | rev
  echo "--------------------------------------------"
fi

node kubeless.js
