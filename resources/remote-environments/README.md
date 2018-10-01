# Remote Environments

## Overview

A Remote Environment (RE) is a representation of an external solution connected to Kyma. The Application Connector manages the traffic, connection, security, and Events of REs. It is a proprietary implementation that consists of four services.
Read the [Application Connector documentation](../../docs/application-connector/docs/001-overview-application-connector.md) for more details regarding the implementation.

## Details

This directory contains the default Remote Environments Helm chart that creates new Remote Environments. 
All of the additional resources required for RE to work are installer by Remote Environment Controller.
A single RE allows to connect a single external solution to Kyma.
