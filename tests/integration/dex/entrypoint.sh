#!/usr/bin/env sh

res=0

echo "Run dex tests"
./dex.test -test.v
res=$((res+$?))

exit ${res}
