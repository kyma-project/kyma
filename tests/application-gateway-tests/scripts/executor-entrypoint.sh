
res=0

echo "Application Gateway tests"

echo "Run proxy tests"
./proxytestsexecutor.test -test.v
res=$((res+$?))

exit ${res}