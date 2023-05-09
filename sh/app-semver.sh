#!/bin/bash

# set semantic version
[[ $(cat FyneApp.toml) =~ ([0-9]+\.[0-9]+\.[0-9]+) ]] && SEMVER="${BASH_REMATCH[0]}"
sed -i -r 's/tag\/v[0-9]+\.[0-9]+\.[0-9]+/tag\/v'${SEMVER}'/g' README.md
sed -i -r 's/tag\/v[0-9]+\.[0-9]+\.[0-9]+/tag\/v'${SEMVER}'/g' docs/README.md
# sed -i -r 's/Version="[0-9]+\.[0-9]+\.[0-9]+"/Version="'${SEMVER}'"/g' sh/app2msi.wsx
echo $SEMVER