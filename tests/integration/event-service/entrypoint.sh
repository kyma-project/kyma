
res=0

echo "Run api tests"
./test.test -test.v
res=$((res+$?))

exit ${res}
