
res=0

echo "Compass Runtime Agent tests"

echo "Run api tests"
./apitests.test -test.v
res=$((res+$?))

exit ${res}