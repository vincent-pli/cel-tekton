# tekton-cel

- Introduce a new CRD named `VariableStore` to record variables. These variables will be treat as `context` when calulate new CEL expression.

```
apiVersion: custom.tekton.dev/v1alpha1
kind: VariableStore
metadata:
  name: example
  namespace: default
spec:
  vars:
  - name: job_priority
    value: high
  - name: alert_enable
    value: yes
```

User could create `VariableStore` previously with some default variables, for example: log_level, time_out .etc

- Introduce a custom controller based on [custom task](https://github.com/tektoncd/community/blob/main/teps/0002-custom-tasks.md)
There are two cased when calculate CEL expression:
1. Without context (when express or condition)
```
apiVersion: tekton.dev/v1alpha1
kind: Run
metadata:
  generateName: celrun-is-red-
spec:
  ref:
    apiVersion: cel.tekton.dev/v1alpha1
    kind: CEL
  params:
    - name: red
      value: "{'blue': '0x000080', 'red': '0xFF0000'}['red']"
    - name: blue
      value: "{'blue': '0x000080', 'red': '0xFF0000'}['blue']"
    - name: is-red
      value: "{'blue': '0x000080', 'red': '0xFF0000'}['red'] == '0xFF0000'"
    - name: is-blue
      value: "{'blue': '0x000080', 'red': '0xFF0000'}['blue'] == '0xFF0000'"
```
2. With context (`user expression`)
```
apiVersion: tekton.dev/v1alpha1
kind: Run
metadata:
  generateName: celrun-with-context-
spec:
  ref:
    apiVersion: custom.tekton.dev/v1alpha1
    kind: VariableStore
    name: example
  params:
    - name: job_severity
      value: "job_priority in ['high', 'normal'] ? 'Sev-1' : 'Sev-2'"
    - name: types
      value: "type(job_severity)"
    - name: alert_enable
      value: "false"      
```
See the `job_severity` leverage `job_priority` which in a context variable in `VariableStore`;
The `types` leverage calculate result of `job_severity`;
`alert_enable` try to reset the variables in `VariableStore`.

After calculate the expression in `params` will be display in the `Result` of the `Run`:
```
spec:
  params:
  - name: job_severity
    value: 'job_priority in [''high'', ''normal''] ? ''Sev-1'' : ''Sev-2'''
  - name: types
    value: type(job_severity)
  - name: alert_enable
    value: "false"
  ref:
    apiVersion: custom.tekton.dev/v1alpha1
    kind: VariableStore
    name: example
  serviceAccountName: default
status:
  completionTime: "2021-04-09T00:48:21Z"
  conditions:
  - lastTransitionTime: "2021-04-09T00:48:21Z"
    message: CEL expressions were evaluated successfully
    reason: EvaluationSuccess
    status: "True"
    type: Succeeded
  extraFields: null
  observedGeneration: 1
  results:
  - name: job_severity
    value: Sev-1
  - name: types
    value: string
  - name: alert_enable
    value: "false"
  startTime: "2021-04-09T00:48:21Z"
  ```
  
  And the new calculated variables will be record to `VariableStore`:
```
apiVersion: custom.tekton.dev/v1alpha1
kind: VariableStore
metadata:
  name: example
  namespace: default
spec:
  vars:
  - name: job_priority
    value: high
  - name: alert_enable
    value: "false"
  - name: job_severity
    value: Sev-1
  - name: types
    value: string
```
