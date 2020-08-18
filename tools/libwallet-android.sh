#!/bin/bash

set -e

repo_root=$(git rev-parse --show-toplevel)

libwallet="$1"
if [[ ! -s "$1" ]]; then
    libwallet="$repo_root/android/apollo/libs/libwallet.aar"
fi

cd "$repo_root/libwallet"

gomobile bind -target=android -o "$libwallet" .

st=$?
echo "rebuilt gomobile with status $? to $libwallet"
exit $st
