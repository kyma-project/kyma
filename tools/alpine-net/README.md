# Alpine-net

## Overview

As part of the Kyma installation, sometimes you need to run shell commands. These shell commands require tools such as `netcat`. For example, you might need the system to wait until a specific host lookup is possible as part of a container startup.
This project provides a docker image, based on alpine having typical network tooling installed. See also the [Dockerfile](Dockerfile)

The alpine-net image has following commands installed:

- net-tools
- bind-tools
- curl
- nmap

## Usage

To build alpine-net locally, call:

```bash
docker build -t alpine-net:latest .
```
