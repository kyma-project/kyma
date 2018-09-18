#!/usr/bin/env bash

NUMBER_OF_PARENTS=$((1 + $(git show -q --format=%P | grep -o ' ' | wc -l)))

if [ $NUMBER_OF_PARENTS -eq 1 ]
then
    echo $(git show -q --format=%P)
elif [ $NUMBER_OF_PARENTS -eq 2 ]
then
    echo $(git rev-parse HEAD^2)
else
    echo There\'s more than two parents of commit.
    exit 1
fi