
res=0

echo "Run api tests"
./apitests.test -test.v
res=$((res+$?))

exit ${res}