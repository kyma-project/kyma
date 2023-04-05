---
title: Ory limitations
---

## Resource configuration

Ory components have the following configuration for resources by default:

| Component          |          | CPU   | Memory |
|--------------------|----------|-------|--------|
| Hydra              | Limits   | 500m  | 256Mi  |
| Hydra              | Requests | 10m   | 128Mi  |
| Oathkeeper         | Limits   | 500m  | 1024Mi |
| Oathkeeper         | Requests | 100m  | 512Mi  |
| Oathkeeper Maester | Limits   | 100m  | 50Mi   |
| Oathkeeper Maester | Requests | 10m   | 20Mi   |

## Autoscaling configuration

The default configuration in terms of autoscaling of Ory components is as follows:

| Component          | Min replicas       | Max replicas       |
|--------------------|--------------------|--------------------|
| Oathkeeper         | 1                  | 3                  |
| Oathkeeper Maester | Same as Oathkeeper | Same as Oathkeeper |
| Hydra              | 1                  | 3                  |

As Oathkeeper Maester is set up as a separate container in the same Pod as Oathkeeper the autoscaling configuration is the same.
