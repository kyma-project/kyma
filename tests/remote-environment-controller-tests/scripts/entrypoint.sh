
res=0

echo "Remote Environment Controller tests"

echo "Run controller tests"
./controller.test
res=$((res+$?))

exit ${res}