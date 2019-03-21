
res=0

echo "Application-Registry tests"

echo "Run api tests"
./apitests.test -test.v
res=$((res+$?))

echo "Run kubernetes tests"
./k8stests.test -test.v
res=$((res+$?))

exit ${res}