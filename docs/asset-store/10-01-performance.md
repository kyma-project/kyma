---
title: Performance
---

This document presents Asset Store performance based on chosen scenarios for particular services. The results are the Asset Store key performance indicators (KPIs).

## Asset Controller

These are the scenarios for the Asset Controller:

- The Asset Controller manages the [Asset CR lifecycle](#details-asset-custom-resource-lifecycle) for a given number of 8MB assets with Markdown files. The process includes metadata extraction and communication with the Webhook service.

| Number of assets | Time |
| 1 |  |
| 10 |  |
| 30 |  |

- The Asset Controller manages the [Asset CR lifecycle](#details-asset-custom-resource-lifecycle) for a given number of 1MB assets with OpenAPI specifications in a single-mode without filtering.

| Number of assets | Time |
| 1 |  |
| 10 |  |
| 30 |  |

## Asset Upload Service

This scenario verifies how many files the Asset Upload Service can upload in a given time.

|Number of files | Time |
|  |  |

## Asset Metadata Service

This scenario verifies from how many files the Asset Metadata Service can extract metadata in a given time.

|Number of files | Time |
|  |  |
