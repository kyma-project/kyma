---
title: Ory limitations
---

## Resource configuration

Ory components have the following configuration for resources by default:

| Component          |          | CPU  | Memory |
|--------------------|----------|------|--------|
| Hydra              | Limits   | 500m | 1Gi    |
| Hydra              | Requests | 250m | 256Mi  |
| Hydra maester      | Limits   | 400m | 1Gi    |
| Hydra maester      | Requests | 10m  | 256Mi  |
| Oathkeeper         | Limits   | 500m | 512Mi  |
| Oathkeeper         | Requests | 100m | 64Mi   |
| Oathkeeper Maester | Limits   | 400m | 1Gi    |
| Oathkeeper Maester | Requests | 10m  | 32Mi   |

## Autoscaling configuration

The default configuration in terms of autoscaling of Ory components is as follows:

| Component          | Min replicas       | Max replicas       |
|--------------------|--------------------|--------------------|
| Oathkeeper         | 3                  | 10                 |
| Oathkeeper Maester | Same as Oathkeeper | Same as Oathkeeper |
| Hydra              | 2                  | 5                  |

As Oathkeeper Maester is set up as a separate container in the same Pod as Oathkeeper the autoscaling configuration is the same.
