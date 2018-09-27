#!/usr/bin/env sh

TEST_EXIT_STATUS=0

for f in *.test; do
    chmod +x ./${f}
    ./${f}
    EXIT_STATUS=$?
    if [ ${EXIT_STATUS} -ne 0 ]; then
        echo "Setting exit status to ${EXIT_STATUS}"
        TEST_EXIT_STATUS=${EXIT_STATUS}
    fi
done

echo "Exiting tests with ${TEST_EXIT_STATUS}"
exit ${TEST_EXIT_STATUS}