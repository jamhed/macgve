#!/bin/sh
kubectl run -it --rm --generator=run-pod/v1 --image=alpine alpine
