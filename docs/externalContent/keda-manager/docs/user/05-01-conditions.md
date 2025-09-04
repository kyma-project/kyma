# Keda Custom Resource Conditions

This document describes the possible conditions of the Keda CR. The conditions make it easier to observe the current state of the Keda module.

The following condition types are used:

- `Installed`
- `Deleted`

| No | CR State   | Condition type | Condition status | Condition reason    | Remark                               |
|----|------------|----------------|------------------|---------------------|--------------------------------------|
| 1  | Ready      | Installed      | true             | Verified            | Server ready                         |
| 2  | Processing | Installed      | unknown          | Initialized         | Initialized                          |
| 3  | Processing | Installed      | unknown          | Verification        | Verification in progress             |
| 4  | Error      | Installed      | false            | ApplyObjError       | Apply object error                   |
| 5  | Error      | Installed      | false            | DeploymentUpdateErr | Deployment update error              |
| 6  | Error      | Installed      | false            | VerificationErr     | Verification error                   |
| 7  | Error      | Installed      | false            | KedaDuplicated      | One instance of Keda is allowed      |
| 8  | Deleting   | Deleted        | unknown          | Deletion            | Deletion in progress                 |
| 9  | Deleting   | Deleted        | true             | Deleted             | Keda module deleted                  |
| 10 | Error      | Deleted        | false            | DeletionErr         | Deletion failed                      |
