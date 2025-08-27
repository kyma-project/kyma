# Serverless Periodically Restarting

## Symptom

When reconciling Git-sourced Functions, Serverless restarts every 10 minutes.

## Cause

Function Controller is polling for changes in referenced Git repositories. If you have a lot of Git-sourced Functions and they were deployed together at approximately the same time, their git sources will be checked out in a synchronized pulse (every 10 minutes). If you happen to reference large repositories (multi-repositories), there will be rhythmical, high demand on CPU and I/O resources necessary to check out repositories. This may cause Function Controller to crash and restart.

## Solution

Avoid using multi-repositories or large repositories to source your git Functions. Using small, dedicated Function repositories decreases the CPU and I/O resources used to check out Git sources, and hence improves the stability of Function Controller.
