#!/bin/bash

PLUGIN_REPO_HOST=$1
PLUGIN_VERSION=$2
COS_API_KEY=$3

node bin/generate-plugin-json.js "${PLUGIN_REPO_HOST}" "${PLUGIN_VERSION}" "${COS_API_KEY}"

cat ./out/plugins.json
echo "Here will will deploy to $1"