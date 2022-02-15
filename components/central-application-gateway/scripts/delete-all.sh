#!/usr/bin/env bash

HOST=51.144.236.138:8081

curl http://$HOST/v1/metadata/services | jq ".[].id" -r > IDs.txt

while read p; do
  echo "$p => $(curl -sIX DELETE http://$HOST/v1/metadata/services/$p | head -n 1)"
  echo
done <IDs.txt

rm IDs.txt
echo "All services removed!"
