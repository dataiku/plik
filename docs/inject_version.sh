#!/usr/bin/env bash

set -e

# Get the version from gen_build_info.sh
VERSION=$(../server/gen_build_info.sh version)

if [[ -z "$VERSION" ]]; then
    echo "Error: Could not determine version"
    exit 1
fi

echo "Injecting version $VERSION into documentation..."

# Update getting-started.md with the current version
sed -i "s|releases/download/[^/]*/plik-[^/]*-linux-amd64.tar.gz|releases/download/$VERSION/plik-$VERSION-linux-amd64.tar.gz|g" guide/getting-started.md
sed -i "s|tar xzvf plik-[^/]*-linux-amd64.tar.gz|tar xzvf plik-$VERSION-linux-amd64.tar.gz|g" guide/getting-started.md
sed -i "s|cd plik-[^/]*/server|cd plik-$VERSION/server|g" guide/getting-started.md

# Update docker.md with the current version
sed -i "s|- \`[0-9.]*\` — Specific version|- \`$VERSION\` — Specific version|g" guide/docker.md

echo "Version injection complete"
