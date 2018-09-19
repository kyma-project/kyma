#!/usr/bin/env bash

NUMBER_OF_PARENTS=$(git show -q --format=%P | wc -w)

if [ $NUMBER_OF_PARENTS -eq 1 ]
then
    echo $(git log -1 --format=%H)
elif [ $NUMBER_OF_PARENTS -eq 2 ]
then
    echo $(git rev-parse HEAD^2)
else
    echo There\'s more than two parents of commit $NUMBER_OF_PARENTS.
    exit 1
fi