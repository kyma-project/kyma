
res=0

echo "Application Operator tests"

echo "Run Service Instance controller tests"
./serviceinstancecontroller.test -test.v
res=$((res+$?))

echo "Run Application controller tests"
./applicationcontroller.test -test.v
res=$((res+$?))

echo "Run application tests"
./applicationtests.test -test.v
res=$((res+$?))

exit ${res}