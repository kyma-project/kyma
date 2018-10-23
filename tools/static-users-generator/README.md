# Static users generator

## Overview

The tool used to configure static users in Dex. Reads users from secrets labelled `"dex-user-config": "true"` and appends them into the Dex config-map.

## Installation

The tool is a Dex init-container used by default in a Kyma installation.

## Development

To build the image of static-users-generator execute:

```bash
docker build -t static-users-generator:latest .
```