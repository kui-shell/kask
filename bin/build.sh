#!/bin/bash

set -e

readonly ROOT_DIR=$(cd $(dirname $(dirname $0)) && pwd)

go build -ldflags "-s -w" -o $ROOT_DIR/out/function-composer .