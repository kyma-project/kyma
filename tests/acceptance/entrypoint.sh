#!/usr/bin/env sh

res=0

### Temporarily disabled
#echo "Run dex tests"
#./dex.test
#res=$((res+$?))

echo "Run servicecatalog tests"
./servicecatalog.test
res=$((res+$?))

echo "Run remote-environment tests"
./remote-environment.test
res=$((res+$?))

exit ${res}
