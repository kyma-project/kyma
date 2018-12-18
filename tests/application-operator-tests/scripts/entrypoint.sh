
res=0

echo "Application Operator tests"

echo "Run controller tests"
./controllertests.test
res=$((res+$?))

exit ${res}