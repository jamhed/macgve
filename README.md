# Govaultenv mutating admission controller

Mutates pods on admission by inserting govaultenv binary and altering pod command and args accordingly.

## Why

To expose vault secrets to applications without altering service definitions.

## How

Annotate pod with `govaultenv.io/authpath` setting value to vault authentication path (e.g. `default@k8s/cluster/namespace`),
and optionally with `govaultenv.io/containers`, to specify comma-separated containers names in pods to apply mutations to.

Pods needs to have command explicitly defined and not to rely on Dockerfile default entry point.

It relies on `govaultenv` docker image, with `govaultenv` and `ca-certificates.crt` files being present in image's
work folder.

## Chart

Please see Helm [chart](https://github.com/jamhed/charts/tree/master/macgve) for deployment details,
and [govaultenv](https://github.com/jamhed/govaultenv) repo.
