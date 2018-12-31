#!/usr/bin/env sh

res=0

echo "Run dex tests"
./dex.test
res=$((res+$?))

echo "Run servicecatalog tests"
./servicecatalog.test
res=$((res+$?))

echo "Run application tests"
./application.test
res=$((res+$?))

exit ${res}
