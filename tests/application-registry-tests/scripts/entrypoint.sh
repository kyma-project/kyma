
res=0

echo "Application-Registry tests"

echo "Run api tests"
./apitests.test
res=$((res+$?))

echo "Run kubernetes tests"
./k8stests.test
res=$((res+$?))

exit ${res}