
res=0

echo "Application Operator tests"

echo "Run controller tests"
./controllertests.test -test.v
res=$((res+$?))

exit ${res}