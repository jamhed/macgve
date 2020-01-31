# Govaultenv mutating admission controller

Mutates pods on admission by inserting govaultenv binary and altering pod command and args accordingly.

## Why

To expose vault secrets to applications without altering service definitions.

## How

Annotate pod with `govaultenv.io/authpath` setting value to vault authentication path (e.g. `default@k8s/cluster/namespace`),
and optionally with `govaultenv.io/containers`, to sepcify containers in pods to apply mutations to.

Pods needs to have command explicitely defined and not to rely on Dockerfile default entrypoint.