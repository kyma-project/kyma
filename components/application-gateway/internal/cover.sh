#!/usr/bin/env bash

#cd internal/
rm coverage*

go test -v -coverprofile coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
open coverage.html
