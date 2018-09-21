
res=0

echo "Remote Environment Controller tests"

echo "Run controller tests"
./controllertests.test
res=$((res+$?))

exit ${res}