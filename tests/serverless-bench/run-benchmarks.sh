#!/usr/bin/env bash

set -o pipefail  # Fail a pipe if any sub-command fails.


ALL_FUNCTIONS=(nodejs14-xs nodejs14-s nodejs14-m nodejs14-l nodejs14-xl python39-s python39-m python39-l python39-xl)

export TEST_NAMESPACE="${NAMESPACE:-serverless-benchmarks}"
export TEST_CONCURRENCY="${CONCURRENCY:-50}"
export TEST_DURATION="${DURATION:-1m}"
export TEST_SPAWN_RATE="${SPAWN_RATE:-50}"


# running benshmarks
for FUNCTION in ${ALL_FUNCTIONS[@]}; do 
    echo "--------------------------------------------------------------------------------"
    echo "Benchmarking function ${FUNCTION} at URL: http://${FUNCTION}.${TEST_NAMESPACE}.svc.cluster.local"
    echo "--------------------------------------------------------------------------------"

    locust --headless --users 50 \
    -r 50 -t 1m \
    -H "http://${FUNCTION}.${TEST_NAMESPACE}.svc.cluster.local" \
    -f /home/locust/locust.py --csv "${FUNCTION}" --logfile /dev/null

done

for FUNCTION in ${ALL_FUNCTIONS[@]}; do 
    echo "--------------------------------------------------------------------------------"
    echo "Collecting Serverless benchmarks..."
    echo "--------------------------------------------------------------------------------"
    export NAME=${FUNCTION}
    mlr --ojson --icsv head -n 1 \
      then put '$Timestamp=system("date --rfc-3339=seconds")' \
      then put '${Function Name} = system("echo $NAME")' \
      then  cut -o \
      -f "Timestamp","Function Name","Request Count","Failure Count","Average Response Time","Requests/s","90%","95%","99%" \
      "${FUNCTION}_stats.csv" | jq '{"serverlessBenchmarkStats": .}' -c
done



