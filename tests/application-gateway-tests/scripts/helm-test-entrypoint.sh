
res=0

echo "Application Gateway tests"

echo "Run proxy tests"
./proxyhelmtests.test -test.v
res=$((res+$?))

exit ${res}