#!/bin/sh

set -e

if [ ! -f "build/env.sh" ]; then
    echo "$0 must be run from the root of the repository."
    exit 2
fi

# Create fake Go workspace if it doesn't exist yet.
workspace="/tmp/aqua/_workspace"
root="$PWD"
aquasrc="$workspace/src/gitlab.com/aquachain"
mkdir -p $aquasrc
if [ ! -L "$aquasrc/aquachain" ]; then
    echo "creating workspace: $aquasrc/aquachain"
    mkdir -p "$aquasrc"
    cd "$aquasrc"
    ln -s $root aquachain
    cd "$root"
fi

# Set up the environment to use the workspace.
GOPATH="$workspace"
export GOPATH

# Run the command inside the workspace.
cd "$aquasrc/aquachain"
PWD="$aquasrc/aquachain"

# Launch the arguments with the configured environment.
exec "$@"
