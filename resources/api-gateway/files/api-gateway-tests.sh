#!/bin/bash
sleep 10

./app.test
exit_code=$?

curl -XPOST http://127.0.0.1:15020/quitquitquit
sleep 5

exit $exit_code
