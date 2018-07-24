#!/bin/bash

DIRS=$(find . -type f -name '*.go' -not -path "./vendor/*")
for d in $DIRS; do goimports -w $d; done