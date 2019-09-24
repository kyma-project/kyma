---
title: Asset Controller
type: Performance
---

This document presents Asset Store performance based on chosen scenarios for particular Asset Store services. The data are the Asset Store key performance indicators (KPIs).

These are the scenarios for the Asset Controller:

- The Asset Controller manages the [Asset CR lifecycle](#details-asset-custom-resource-lifecycle) for a given number of 8MB assets with Markdown files. The process includes metadata extraction and communication with the Webhook service.

| Number of assets | Time |
|------------------|------|
| 1 |  |
| 10 |  |
| 30 |  |

- The Asset Controller manages the [Asset CR lifecycle](#details-asset-custom-resource-lifecycle) for a given number of 1MB assets with OpenAPI specifications in a single-mode without filtering.

| Number of assets | Time |
|------------------|------|
| 1 |  |
| 10 |  |
| 30 |  |
