#!/bin/sh
# Verify the actual artifact, not just the build environment.
set -eu

binary=${1:?usage: check-macos-build.sh BINARY ARCH CGO}
arch=${2:?missing architecture}
cgo=${3:?missing expected CGO value}
target=${MACOSX_DEPLOYMENT_TARGET:-12.0}

metadata=$(go version -m "$binary")
for setting in "GOOS=darwin" "GOARCH=$arch" "CGO_ENABLED=$cgo"; do
    if ! printf '%s\n' "$metadata" | awk -v expected="$setting" '$1 == "build" && $2 == expected { found = 1 } END { exit !found }'; then
        echo "Unexpected build settings for $binary: missing $setting" >&2
        exit 1
    fi
done

minimum=$(otool -l "$binary" | awk '$1 == "minos" { print $2 }')
if [ "$minimum" != "$target" ]; then
    echo "$binary requires macOS $minimum; expected $target" >&2
    exit 1
fi
echo "Verified $binary: darwin/$arch, CGO=$cgo, macOS $minimum"
