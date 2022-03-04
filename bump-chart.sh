#!/bin/bash -ae
CHART=$1
VERSION=$(yq .version $CHART)
MINOR=$(echo $VERSION | sed -E 's|^.*\.([0-9]+)$|\1|')
MAJOR=$(echo $VERSION | sed -E 's|^v(.*)\.[0-9]+$|\1|')
NEWVER="$MAJOR.$((MINOR+1))"
yq -i '.version=env(NEWVER)' $CHART
