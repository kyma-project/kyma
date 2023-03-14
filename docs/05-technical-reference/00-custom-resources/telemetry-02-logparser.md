---
title: Telemetry - LogParser
---

The `logparser.telemetry.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to define a custom log parser in Kyma. To get the up-to-date CRD and show the output in the YAML format, run this command:

```bash
kubectl get crd logparser.telemetry.kyma-project.io -o yaml
```

## Sample custom resource

The following LogParser object defines a parser which can parse a log line like: `{"data":"100 0.5 true This is example"}`.

```yaml
apiVersion: telemetry.kyma-project.io/v1alpha1
kind: LogParser
metadata:
  name: my-regex-parser
spec:
  parser: |
    Format regex
    Regex ^(?<INT>[^ ]+) (?<FLOAT>[^ ]+) (?<BOOL>[^ ]+) (?<STRING>.+)$
```

For further LogParser examples, see the [samples](https://github.com/kyma-project/telemetry-manager/tree/main/config/samples) directory.

## Custom resource parameters

For details, see the [LogParser specification file](https://github.com/kyma-project/telemetry-manager/blob/main/apis/telemetry/v1alpha1/logparser_types.go).

<!-- The table below was generated automatically -->
<!-- Some special tags (html comments) are at the end of lines due to markdown requirements. -->
<!-- The content between "TABLE-START" and "TABLE-END" will be replaced -->

<!-- SKIP-WITH-ANCESTORS spec.template -->

<!-- TABLE-START -->
<!-- LogParser -->
| Parameter         | Description                                   |
| ---------------------------------------- | ---------|
| **spec.parser** | Configures a user defined Fluent Bit parser to be applied to the logs. |
| **status.conditions** | LogParserCondition contains details for the current condition of this LogParser |
| **status.conditions.lastTransitionTime** |  |
| **status.conditions.reason** |  |
| **status.conditions.type** |  |<!-- TABLE-END -->
