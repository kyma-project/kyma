# LogParser

The `logparser.telemetry.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to define a custom log parser in Kyma. To get the current CRD and show the output in the YAML format, run this command:

```bash
kubectl get crd logparser.telemetry.kyma-project.io -o yaml
```

## Sample Custom Resource

The following LogParser object defines a parser which can parse a log line like: `{"data":"100 0.5 true This is example"}`.

```yaml
apiVersion: telemetry.kyma-project.io/v1alpha1
kind: LogParser
metadata:
  name: my-regex-parser
  generation: 1
spec:
  parser: |
    Format regex
    Regex ^(?<INT>[^ ]+) (?<FLOAT>[^ ]+) (?<BOOL>[^ ]+) (?<STRING>.+)$
status:
  conditions:
  - lastTransitionTime: "2024-02-29T01:27:08Z"
    message: Fluent Bit DaemonSet is ready
    observedGeneration: 1
    reason: DaemonSetReady
    status: "True"
    type: AgentHealthy
```

For further examples, see the [samples](https://github.com/kyma-project/telemetry-manager/tree/main/config/samples) directory.

## Custom Resource Parameters

For details, see the [LogParser specification file](https://github.com/kyma-project/telemetry-manager/blob/main/apis/telemetry/v1alpha1/logparser_types.go).

<!-- The table below was generated automatically -->
<!-- Some special tags (html comments) are at the end of lines due to markdown requirements. -->
<!-- The content between "TABLE-START" and "TABLE-END" will be replaced -->

<!-- TABLE-START -->
### LogParser.telemetry.kyma-project.io/v1alpha1

>**CAUTION**: The LogParser API is deprecated. Instead, log in JSON format and use the JSON parsing feature of the LogPipeline

**Spec:**

| Parameter | Type | Description |
| ---- | ----------- | ---- |
| **parser**  | string | [Fluent Bit Parsers](https://docs.fluentbit.io/manual/pipeline/parsers). The parser specified here has no effect until it is referenced by a [Pod annotation](https://docs.fluentbit.io/manual/pipeline/filters/kubernetes#kubernetes-annotations) on your workload or by a [Parser Filter](https://docs.fluentbit.io/manual/pipeline/filters/parser) defined in a pipeline's filters section. |

**Status:**

| Parameter | Type | Description |
| ---- | ----------- | ---- |
| **conditions**  | \[\]object | An array of conditions describing the status of the parser. |
| **conditions.&#x200b;lastTransitionTime** (required) | string | lastTransitionTime is the last time the condition transitioned from one status to another. This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable. |
| **conditions.&#x200b;message** (required) | string | message is a human readable message indicating details about the transition. This may be an empty string. |
| **conditions.&#x200b;observedGeneration**  | integer | observedGeneration represents the .metadata.generation that the condition was set based upon. For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date with respect to the current state of the instance. |
| **conditions.&#x200b;reason** (required) | string | reason contains a programmatic identifier indicating the reason for the condition's last transition. Producers of specific condition types may define expected values and meanings for this field, and whether the values are considered a guaranteed API. The value should be a CamelCase string. This field may not be empty. |
| **conditions.&#x200b;status** (required) | string | status of the condition, one of True, False, Unknown. |
| **conditions.&#x200b;type** (required) | string | type of condition in CamelCase or in foo.example.com/CamelCase. |

<!-- TABLE-END -->

### LogParser Status

The status of the LogParser is determined by the condition type `AgentHealthy`:

| Condition Type | Condition Status | Condition Reason  | Condition Message                 |
|----------------|------------------|-------------------|-----------------------------------|
| AgentHealthy   | True             | DaemonSetReady    | Fluent Bit DaemonSet is ready     |
| AgentHealthy   | False            | DaemonSetNotReady | Fluent Bit DaemonSet is not ready |
