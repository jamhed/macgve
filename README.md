# Govaultenv Kubernetes injector

Injects [govaultenv](https://github.com/jamhed/govaultenv) binary to annotated kubernetes pods.

## Quick start

```sh
helm repo add gve https://jamhed.github.io/macgve
```

## How

Annotate pod with `govaultenv.io/authpath` setting value to vault authentication
path (e.g. `default@kubernetes/cluster/namespace`), and optionally with `govaultenv.io/containers`,
to specify comma-separated containers names in pods to apply mutations to.

Pods needs to have command explicitly defined and not to rely on Dockerfile default entry point.
