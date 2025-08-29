# Supported API Gateway Configurations

## Status

Accepted

## Context

While having different configuration options for APIRules, we specify which combinations of **path** + **accessStrategies.handler** are allowed for the API Gateway component. Additionally, we will provide validation based on this.

## Decision

After having a meeting today and discussed the issue within the team, we decided on the following:

### Handlers Matrix Based on a Rule with 1 Path and any 2+ Handlers

| First/Second handler | `allow` | `noop` | `oauth2` | `jwt` (ory) | `jwt` (istio) |
|:---:|:---:|:---:|:---:|:---:|:---:|
| `allow` | ![](https://placehold.co/15x15/f03c15/f03c15.png) | ![](https://placehold.co/15x15/f03c15/f03c15.png) | ![](https://placehold.co/15x15/f03c15/f03c15.png) | ![](https://placehold.co/15x15/f03c15/f03c15.png) | ![](https://placehold.co/15x15/f03c15/f03c15.png) |
| `noop` | ![](https://placehold.co/15x15/f03c15/f03c15.png) | ![](https://placehold.co/15x15/00FF00/00FF00.png) | ![](https://placehold.co/15x15/00FF00/00FF00.png) | ![](https://placehold.co/15x15/00FF00/00FF00.png) | ![](https://placehold.co/15x15/f03c15/f03c15.png) |
| `oauth2` | ![](https://placehold.co/15x15/f03c15/f03c15.png) | ![](https://placehold.co/15x15/00FF00/00FF00.png) | ![](https://placehold.co/15x15/00FF00/00FF00.png) | ![](https://placehold.co/15x15/00FF00/00FF00.png) | ![](https://placehold.co/15x15/f03c15/f03c15.png) |
| `jwt` (ory) | ![](https://placehold.co/15x15/f03c15/f03c15.png) | ![](https://placehold.co/15x15/00FF00/00FF00.png) | ![](https://placehold.co/15x15/00FF00/00FF00.png) | ![](https://placehold.co/15x15/00FF00/00FF00.png) | ![](https://placehold.co/15x15/f03c15/f03c15.png) |
| `jwt` (istio) | ![](https://placehold.co/15x15/f03c15/f03c15.png) | ![](https://placehold.co/15x15/f03c15/f03c15.png) | ![](https://placehold.co/15x15/f03c15/f03c15.png) | ![](https://placehold.co/15x15/f03c15/f03c15.png) | ![](https://placehold.co/15x15/f03c15/f03c15.png) |

### Handlers Matrix Based on a Rule with any 2+ Paths and a Handler

| First/Second path | `allow` | `noop` | `oauth2` | `jwt` (ory) | `jwt` (istio) |
|:---:|:---:|:---:|:---:|:---:|:---:|
| `allow` | ![](https://placehold.co/15x15/00FF00/00FF00.png) | ![](https://placehold.co/15x15/00FF00/00FF00.png) | ![](https://placehold.co/15x15/00FF00/00FF00.png) | ![](https://placehold.co/15x15/00FF00/00FF00.png) | ![](https://placehold.co/15x15/00FF00/00FF00.png) |
| `noop` | ![](https://placehold.co/15x15/00FF00/00FF00.png) | ![](https://placehold.co/15x15/00FF00/00FF00.png) | ![](https://placehold.co/15x15/00FF00/00FF00.png) | ![](https://placehold.co/15x15/00FF00/00FF00.png) | ![](https://placehold.co/15x15/00FF00/00FF00.png) |
| `oauth2` | ![](https://placehold.co/15x15/00FF00/00FF00.png) | ![](https://placehold.co/15x15/00FF00/00FF00.png) | ![](https://placehold.co/15x15/00FF00/00FF00.png) | ![](https://placehold.co/15x15/00FF00/00FF00.png) | ![](https://placehold.co/15x15/00FF00/00FF00.png) |
| `jwt (ory)` | ![](https://placehold.co/15x15/00FF00/00FF00.png) | ![](https://placehold.co/15x15/00FF00/00FF00.png) | ![](https://placehold.co/15x15/00FF00/00FF00.png) | ![](https://placehold.co/15x15/00FF00/00FF00.png) | ![](https://placehold.co/15x15/f03c15/f03c15.png) |
| `jwt (istio)` | ![](https://placehold.co/15x15/00FF00/00FF00.png) | ![](https://placehold.co/15x15/00FF00/00FF00.png) | ![](https://placehold.co/15x15/00FF00/00FF00.png) | ![](https://placehold.co/15x15/f03c15/f03c15.png) | ![](https://placehold.co/15x15/00FF00/00FF00.png) |

### Legend

| Type | Description |
|:---:|:---|
| ![](https://placehold.co/15x15/00FF00/00FF00.png) | Supported |
| ![](https://placehold.co/15x15/f03c15/f03c15.png) | Not supported |

## Consequences

Having specified supported configurations, we can improve our unit test coverage in API Gateway and add the needed validation for APIRules.