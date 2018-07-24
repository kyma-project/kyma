res=0

echo "Run api tests"
./apitests.test
res=$((res+$?))

exit ${res}