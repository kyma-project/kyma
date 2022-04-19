---
title: Serverless periodically restarting
---


## Symptom

Serverless restarting every 10 minutes when reconciling git-sourced functions.

## Cause

Function controller is polling for changes in referenced git repositories. If you have a lot of git-sourced functions and they were deployed together in approximately the same time their git sources will be checked-out in a synchronised pulse (every 10 minutes). If you happen to reference large repositories (multi-repositories) there will be rhythmical high demand on cpu and I/O resources necessary to checkout repositories that may cause function controller to crash and restart.

## Remedy

Avoid using multi-repositories or large repositories to source your git functions. Using small or dedicated function repositories will decrease cpu and I/O  resources used to checkout git sources and hence assure better stability of functions controller.
