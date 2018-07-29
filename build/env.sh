#!/bin/sh

set -e

if [ ! -f "build/env.sh" ]; then
    echo "$0 must be run from the root of the repository."
    exit 2
fi

# Create fake Go workspace if it doesn't exist yet.
workspace="/tmp/aqua/_workspace"
root="$PWD"
ethdir="$workspace/src/gitlab.com/aquachain"
mkdir -p $ethdir
if [ ! -L "$ethdir/aquachain" ]; then
    echo "creating workspace: $ethdir/aquachain"
    mkdir -p "$ethdir"
    cd "$ethdir"
    ln -s $root aquachain
    cd "$root"
fi

# Set up the environment to use the workspace.
GOPATH="$workspace"
export GOPATH

# Run the command inside the workspace.
cd "$ethdir/aquachain"
PWD="$ethdir/aquachain"

# Launch the arguments with the configured environment.
exec "$@"
