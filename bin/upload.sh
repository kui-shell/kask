#!/usr/bin/env bash

readonly ROOT_DIR=$(cd $(dirname $(dirname $0)) && pwd)
readonly OUT_DIR=$ROOT_DIR/out

VERSION=$1
IAM_TOKEN=$2
URL="https://s3-api.us-geo.objectstorage.softlayer.net/shelldist/dist/$VERSION"

for file in "$OUT_DIR"/*
do
  if [[ $file != "$OUT_DIR/checksum" && $file != "OUT_DIR/plugins.json" ]]; then
    binary=${file##*/}
    cmd=`curl -v -X PUT "$URL/$binary-$VERSION" \
        -H "x-amz-acl: public-read" \
        -H "Authorization: Bearer $IAM_TOKEN" \
        -H "Content-Type: application/octet-stream" \
        --data-binary "@$file"`
     echo $cmd
  fi
done