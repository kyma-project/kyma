#!/bin/sh
while ! curl -Ik https://$KUBERNETES_SERVICE_HOST:$KUBERNETES_SERVICE_PORT_HTTPS >/dev/null 2>/dev/null
do
    sleep 1
done

/restore.test -test.v
