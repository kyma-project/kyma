#!/usr/bin/env bash

NUMBER_OF_PARENTS=$(git show -q --format=%P | wc -w)

if [ $NUMBER_OF_PARENTS -eq 1 ]
then
    echo $(git rev-parse HEAD)
elif [ $NUMBER_OF_PARENTS -eq 2 ]
then
    echo $(git rev-parse HEAD^2)
else
    echo Can\'t resolve valid commit hash, there are $NUMBER_OF_PARENTS parents of HEAD \(should be 1 or 2\).
    exit 1
fi