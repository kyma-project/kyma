#!/usr/bin/env sh
set -e
set -o pipefail

configmap_path="/mnt/cfg-cm"
out_path="/mnt/cfg-vol/"

if [ ! -d ${configmap_path} ]; then
    echo "${configmap_path} not found. done."
    exit 0
fi

if [ -d ${out_path} ]; then
    cd ${out_path}
    cp -Lv ${configmap_path}/*.conf .

    GCFG="gnatsd.conf"
    echo "authorization {" >> ${GCFG}
    echo "  users = [" >> ${GCFG}
    echo "    {user: ${STAN_USERNAME}, password: ${STAN_PASSWD}}" >> ${GCFG}
    echo "    {user: ${EB_USERNAME}, password: ${EB_PASSWD}}" >> ${GCFG}
    echo "    {user: ${KN_USERNAME}, password: ${KN_PASSWD}}" >> ${GCFG}
    echo "  ]" >> ${GCFG}
    echo "}" >> ${GCFG}
    
    cat ${GCFG}
    echo "${out_path} prepared. done."
else 
    echo "${out_path} not found. done."
fi