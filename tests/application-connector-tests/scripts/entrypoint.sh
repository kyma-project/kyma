
res=0

echo "Application Operator tests"

echo "Run Application Access tests"
./applicationaccess.test -test.v
res=$((res+$?))

exit ${res}