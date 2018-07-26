#!/usr/bin/env bash

readonly ROOT_DIR=$(cd $(dirname $(dirname $0)) && pwd)
readonly OUT_DIR=$ROOT_DIR/out

VERSION=$1
IAM_TOKEN=$2
HOST="s3-api.us-geo.objectstorage.softlayer.net"

for file in "$OUT_DIR"/*
do
  if [[ $file != "$OUT_DIR/checksum" && $file != "OUT_DIR/plugins.json" ]]; then
    binary=${file##*/}
    cmd=`curl -v -X PUT "https://$HOST/shelldist/dist/$VERSION/$binary-$VERSION" \
        -H "x-amz-acl: public-read" \
        -H "Authorization: $IAM_TOKEN" \
        -H "Content-Type: application/octet-stream" \
        --data-binary "@$file"`
     $cmd
  fi
done