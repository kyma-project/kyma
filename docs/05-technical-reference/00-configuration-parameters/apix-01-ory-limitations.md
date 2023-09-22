---
title: Ory limitations
---

## Resource configuration

By default, the Ory components' resources have the following configuration:

| Component          |          | CPU  | Memory |
|--------------------|----------|------|--------|
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

Oathkeeper Maester is set up as a separate container in the same Pod as Oathkeeper. Because of that, their autoscaling configuration is similar.
