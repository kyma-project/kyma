#!/usr/bin/env bash

set -o pipefail  # Fail a pipe if any sub-command fails.


ALL_FUNCTIONS="nodejs16-xs nodejs16-s nodejs16-m nodejs16-l nodejs16-xl nodejs14-xs nodejs14-s nodejs14-m nodejs14-l nodejs14-xl python39-s python39-m python39-l python39-xl"

export TEST_NAMESPACE="${NAMESPACE:-serverless-benchmarks}"
export TEST_CONCURRENCY="${CONCURRENCY:-50}"
export TEST_DURATION="${DURATION:-1m}"
export TEST_SPAWN_RATE="${SPAWN_RATE:-50}"
export TEST_USE_LOCAL_MYSQL="${USE_LOCAL_MYSQL:-false}"
export TEST_LOCAL_MYSQL_HOST="${LOCAL_MYSQL_HOST:-mysql-bench-db.default.svc.cluster.local}"
export TEST_LOCAL_MYSQL_USER="${LOCAL_MYSQL_USER:-root}"
export TEST_LOCAL_MYSQL_PASS="${LOCAL_MYSQL_PASS:-secret}"
export TEST_LOCAL_MYSQL_DB="${LOCAL_MYSQL_DB:-mysql_bench_db}"

export TEST_CUSTOM_FUNCTIONS="${CUSTOM_FUNCTIONS:-""}"

if [ ! -z "${TEST_CUSTOM_FUNCTIONS}" ]; then
    ALL_FUNCTIONS="${TEST_CUSTOM_FUNCTIONS}"
fi

    echo "--------------------------------------------------------------------------------"
    echo "Benchmarking functions: ${ALL_FUNCTIONS}"
    echo "Using local MySQL backend: ${TEST_USE_LOCAL_MYSQL}"
    echo "--------------------------------------------------------------------------------"
# running benchmarks
for FUNCTION in ${ALL_FUNCTIONS}; do 
    echo "--------------------------------------------------------------------------------"
    echo "Benchmarking function ${FUNCTION} at URL: http://${FUNCTION}.${TEST_NAMESPACE}.svc.cluster.local"
    echo "--------------------------------------------------------------------------------"

    locust --headless --users ${TEST_CONCURRENCY} \
    -r ${SPAWN_RATE} -t ${TEST_DURATION} \
    -H "http://${FUNCTION}.${TEST_NAMESPACE}.svc.cluster.local" \
    -f /home/locust/locust.py --csv "${FUNCTION}" --logfile /dev/null

done

    echo "--------------------------------------------------------------------------------"
    echo "Collecting Serverless benchmarks..."
    echo "--------------------------------------------------------------------------------"
for FUNCTION in ${ALL_FUNCTIONS}; do 
    export NAME=${FUNCTION}
    if [ "${TEST_USE_LOCAL_MYSQL}" = "false" ]; then
        mlr --ojson --icsv head -n 1 \
        then put '$Timestamp=system("date --rfc-3339=seconds")' \
        then put '${Function Name} = system("echo $NAME")' \
        then  cut -o \
        -f "Timestamp","Function Name","Request Count","Failure Count","Average Response Time","Requests/s","90%","95%","99%" \
        "${FUNCTION}_stats.csv" | jq '{"serverlessBenchmarkStats": .}' -c
    else
        mlr --ocsv --headerless-csv-output --icsv head -n 1 \
        then put '$Timestamp=system("date --rfc-3339=seconds")' \
        then put '${Function Name} = system("echo $NAME")' \
        then  cut -o \
        -f "Function Name","Request Count","Failure Count","Average Response Time","Requests/s","90%","95%","99%","Timestamp" \
        "${FUNCTION}_stats.csv" >> local_mysql_data.csv
    fi
done

if [ "${USE_LOCAL_MYSQL}" = "true" ]; then
    echo "--------------------------------------------------------------------------------"
    echo "Importing data into local MySQL database..."
    echo "--------------------------------------------------------------------------------"
    mysql -u "${TEST_LOCAL_MYSQL_USER}" \
        -p"${TEST_LOCAL_MYSQL_PASS}" \
        -h "${TEST_LOCAL_MYSQL_HOST}" \
        -e "CREATE DATABASE IF NOT EXISTS "${TEST_LOCAL_MYSQL_DB}"; \
        USE "${TEST_LOCAL_MYSQL_DB}"; \
        CREATE TABLE IF NOT EXISTS serverless_bench (function_name VARCHAR(128),\
        request_count INT, failure_count INT, avg_response FLOAT, rqs FLOAT, percentile90 FLOAT,percentile95 FLOAT,percentile99 FLOAT, timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP);"

    mysql -u "${TEST_LOCAL_MYSQL_USER}" \
        -p"${TEST_LOCAL_MYSQL_PASS}" \
        -h "${TEST_LOCAL_MYSQL_HOST}" \
        "$TEST_LOCAL_MYSQL_DB" \
        -e "LOAD DATA LOCAL INFILE '/home/locust/local_mysql_data.csv' INTO TABLE serverless_bench FIELDS TERMINATED BY ',' LINES TERMINATED BY '\n';"
fi
