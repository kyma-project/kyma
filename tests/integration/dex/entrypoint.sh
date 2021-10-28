#!/usr/bin/env sh

res=0

echo "Run dex tests"
/tests/dex.test -test.v
res=$((res+$?))

exit ${res}
