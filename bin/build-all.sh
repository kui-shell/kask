#!/bin/bash

set -ex

readonly ROOT_DIR=$(cd $(dirname $(dirname $0)) && pwd)
readonly OUT_DIR=$ROOT_DIR/out

rm -rf $OUT_DIR

build() {
    os=$1
    arch=$2
    GOARCH=$arch GOOS=$os $ROOT_DIR/bin/build.sh

    nf="function-composer-$os-$arch"
    if [ $os == "windows" ]; then
        nf="$nf.exe"
    fi
    mv $OUT_DIR/function-composer "$OUT_DIR/$nf"
}

build windows amd64
build windows 386
# disable CGO for Linux
CGO_ENABLED=0 build linux amd64
CGO_ENABLED=0 build linux 386
#CGO_ENABLED=0 build linux ppc64le
build darwin amd64

shasum $OUT_DIR/*
shasum $OUT_DIR/* > $OUT_DIR/checksums