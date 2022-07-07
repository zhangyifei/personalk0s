#!/bin/bash -e

# get build version using the latest git tag in build branch
# if BUILD_TAG is set, use BUILD_TAG regardless of git tag
# if any checkout files, or no tag found, use commitid-dirty
if [ -n "$(git status --porcelain --untracked-files=no)" ]; then
    DIRTY="-dirty"
fi

COMMIT=$(git rev-parse --short HEAD)
GIT_TAG=${BUILD_TAG:-$(git tag -l --contains HEAD | head -n 1)}

if [[ -z "$DIRTY" && -n "$GIT_TAG" ]]; then
    VERSION=$GIT_TAG
else
    VERSION="${COMMIT}${DIRTY}"
fi

# start to build. 
# if CROSS is set, build for three platforms. otherwise just build for current.
cd $(dirname $0)/..

if [ -n "$CROSS" ]; then
    OS_PLATFORM_ARG=(linux windows darwin)
    OS_ARCH_ARG=(amd64 arm64)
    rm -rf build/bin
    mkdir -p build/bin
    for OS in ${OS_PLATFORM_ARG[@]}; do
        for ARCH in ${OS_ARCH_ARG[@]}; do
            OUTPUT_BIN="build/bin/$OS/$ARCH/eke"
            if test "$OS" = "windows"; then
                OUTPUT_BIN="${OUTPUT_BIN}.exe"
            fi
            echo "Building binary for $OS/$ARCH..."
            GOARCH=$ARCH GOOS=$OS CGO_ENABLED=0 go build -ldflags="-s -w -X eke/pkg/build.Version=$VERSION" -o ${OUTPUT_BIN} ./
        done
    done
else
    CGO_ENABLED=0 go build -ldflags="-s -w -X eke/pkg/build.Version=$VERSION" -o build/bin/eke ./
fi

# https://www.arp242.net/static-go.html explains well
