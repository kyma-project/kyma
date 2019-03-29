res=0

echo "Connection Token Handler tests"

echo "Run kubernetes tests"
./k8stests.test
res=$((res+$?))

exit ${res}