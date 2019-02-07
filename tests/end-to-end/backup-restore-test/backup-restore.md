# Backup and restore E2E tests

## Overview

This project contains end-to-end tests that it runs as part of the  Kyma on Google Cloud Platform installation. The tests are written in Go. 

- Create resources, such as functions on the Kyma cluster.
- Backup the cluster.
- Recreate the cluster.
- Restore data and check if the restored resources work as expected.

## Usage

