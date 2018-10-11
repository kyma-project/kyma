# Static users generator

## Overview

The tool used to configure static users in Dex. Reads users from secrets labelled `"dex-user-config": "true"` and put them into the Dex config-map.

## Installation

The tool should be used as Dex init-container. Dex installation in Kyma uses it be default.

## Usage

To create the static user please create a secret with label `"dex-user-config": "true"`.

> The secrets should exist before the dex installation, because the init-container runs only in the time of Dex pod creation.
> To add users after the Dex pod is created, you can restart the Dex pod. That will cause creation of another Dex pod, with configmap updated.

The secret data should contain three fields, all three encoded with base64:
*email* - the email which will be used to log into the console
*username* - the username which will be displayed in console
*password* - the password which will be used to log into the console (there are no specific prerequisites for a password, however it is suggested to make it at least 8 characters long)

If any of these fields will be missing the user won't be created.