
res=0

echo "Application-connector tests"

echo "Run api tests"
./apitests.test
res=$((res+$?))

echo "Run kubernetes tests"
./k8stests.test
res=$((res+$?))

exit ${res}