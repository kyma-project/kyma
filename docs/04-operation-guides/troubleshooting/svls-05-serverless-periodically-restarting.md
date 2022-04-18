---
title: Serverless periodically restarting
---


## Symptom

Serverless restarting every 10 minutes when checking for function changes.

## Cause

Git functions referencing large repositories cost a lot of resources to pull repositories and check for function changes. That causes serverless to restart.

## Remedy

Avoid using multi-repositories or large repositories to source your git functions. Using small or dedicated function repositories will decrease resources used and prevent serverless from restarting.

