# Backup and restore E2E tests

## Overview

This project contains the end to end tests that it runs as part of the Kyma on Google Cloud Platform. The tests is written in Go. 

- Create resources on Kyma Cluster (e.g. functions).
- Backup Cluster.
- Recreate Cluster.
- Restore Data and check if the restored resources are working like expected.

## Usage

