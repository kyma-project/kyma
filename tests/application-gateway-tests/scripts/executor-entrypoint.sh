
res=0

echo "Application Gateway tests"

echo "Run proxy tests"
./proxytestsexecutor.test
res=$((res+$?))

exit ${res}