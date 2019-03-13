
res=0

echo "Application Gateway tests"

echo "Run proxy tests"
./proxyhelmtests.test
res=$((res+$?))

exit ${res}