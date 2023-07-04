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

<!-- TABLE-START -->
### LogParser.telemetry.kyma-project.io/v1alpha1

**Spec:**

| Parameter | Type | Description |
| ---- | ----------- | ---- |
| **parser**  | string | [Fluent Bit Parsers](https://docs.fluentbit.io/manual/pipeline/parsers). The parser specified here has no effect until it is referenced by a [Pod annotation](https://docs.fluentbit.io/manual/pipeline/filters/kubernetes#kubernetes-annotations) on your workload or by a [Parser Filter](https://docs.fluentbit.io/manual/pipeline/filters/parser) defined in a pipeline's filters section. |

**Status:**

| Parameter | Type | Description |
| ---- | ----------- | ---- |
| **conditions**  | \[\]object | An array of conditions describing the status of the parser. |
| **conditions.&#x200b;lastTransitionTime**  | string | An array of conditions describing the status of the parser. |
| **conditions.&#x200b;reason**  | string | An array of conditions describing the status of the parser. |
| **conditions.&#x200b;type**  | string | The possible transition types are:<br>- `Running`: The parser is ready and usable.<br>- `Pending`: The parser is being activated. |

<!-- TABLE-END -->
