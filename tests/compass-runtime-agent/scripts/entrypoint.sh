
res=0

echo "Compass Runtime Agent tests"

echo "Run tests"
./test.test -test.v
res=$((res+$?))

exit ${res}