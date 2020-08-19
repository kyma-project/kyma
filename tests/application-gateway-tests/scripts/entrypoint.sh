
res=0

echo "Application Gateway tests"

echo "Run Gateway tests"
./tests.test -test.v
res=$((res+$?))

exit ${res}