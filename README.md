# pod-restarts-check

The `pod-restarts-check` looks for `Warning` events with reason `BackOff` and reports failure when the count exceeds a configured threshold.

## Configuration

Set these environment variables in the `HealthCheck` spec:

- `POD_NAMESPACE` (optional): namespace to inspect. Defaults to all namespaces.
- `MAX_FAILURES_ALLOWED` (optional): maximum allowed `BackOff` event count. Defaults to `10`.
- `KUBECONFIG` (optional): explicit kubeconfig path for local development.

When targeting all namespaces, the service account needs cluster-wide permissions for pods and events.

## Build

- `just build` builds the container image locally.
- `just test` runs unit tests.
- `just binary` builds the binary in `bin/`.

## Example HealthCheck

Apply the example below or the provided `healthcheck.yaml`:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: pod-restarts-check
  namespace: kuberhealthy
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: pod-restarts-check
  namespace: kube-system
rules:
  - apiGroups: [""]
    resources: ["events", "pods"]
    verbs: ["get", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: pod-restarts-check
  namespace: kube-system
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: pod-restarts-check
subjects:
  - kind: ServiceAccount
    name: pod-restarts-check
    namespace: kuberhealthy
---
apiVersion: kuberhealthy.github.io/v2
kind: HealthCheck
metadata:
  name: pod-restarts
  namespace: kuberhealthy
spec:
  runInterval: 5m
  timeout: 10m
  podSpec:
    spec:
      serviceAccountName: pod-restarts-check
      containers:
        - name: pod-restarts
          image: kuberhealthy/pod-restarts-check:sha-<short-sha>
          imagePullPolicy: IfNotPresent
          env:
            - name: POD_NAMESPACE
              value: "kube-system"
            - name: MAX_FAILURES_ALLOWED
              value: "10"
          resources:
            requests:
              cpu: 10m
              memory: 50Mi
      restartPolicy: Never
```
