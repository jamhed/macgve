# Govaultenv Kubernetes injector

Injects [govaultenv](https://github.com/jamhed/govaultenv) binary to annotated kubernetes pods.

## Quick start

```sh
helm repo add gve https://jamhed.github.io/macgve
helm -n macgve install --create-namespace macgve gve/macgve --set macgve.vaultAddr=https://...
```

## How

Annotate pod with `govaultenv.io/authpath` annotation with value of vault authentication
path (e.g. `kubernetes`), and optionally with `govaultenv.io/containers`,
with value set to  comma-separated containers names in pods to apply mutations to.

It's also possible to annotate the pod's namespace with the same `govaultenv.io/authpath` annotation.

`macgve` uses service account name as the vault role to use to authentifice to vault, e.g. `default`.

Pods needs to have command explicitly defined and not to rely on Dockerfile default entry point.

## Development

```sh
skaffold dev --status-check=false
```
