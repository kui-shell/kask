#!/usr/bin/env bash

readonly ROOT_DIR=$(cd $(dirname $(dirname $0)) && pwd)
readonly OUT_DIR=$ROOT_DIR/out

VERSION=$1
URL=$2
IAM_TOKEN=$3
URL="$2"

for file in "$OUT_DIR"/*
do
  if [[ $file != "$OUT_DIR/checksums" && $file != "$OUT_DIR/plugins.json" ]]; then
    binary=${file##*/}
    cmd=`curl -v -X PUT "$URL/$VERSION/$binary-$VERSION" \
        -H "x-amz-acl: public-read" \
        -H "Authorization: Bearer $IAM_TOKEN" \
        -H "Content-Type: application/octet-stream" \
        --data-binary "@$file"`
     echo $cmd
  fi
done