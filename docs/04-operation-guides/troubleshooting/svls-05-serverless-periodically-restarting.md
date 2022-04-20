---
title: Serverless periodically restarting
---


## Symptom

Serverless restarting every 10 minutes when reconciling git-sourced Functions.

## Cause

Function controller is polling for changes in referenced git repositories. If you have a lot of git-sourced Functions and they were deployed together at approximately the same time their git sources will be checked out in a synchronised pulse (every 10 minutes). If you happen to reference large repositories (multi-repositories), there will be rhythmical high demand on cpu and I/O resources necessary to checkout repositories. This may cause Function controller to crash and restart.

## Remedy

Avoid using multi-repositories or large repositories to source your git Functions. Using small, dedicated Function repositories will decreases the cpu and I/O  resources used to check out git sources, and hence improve the stability of Function controller.
