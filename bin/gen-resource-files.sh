#!/bin/bash

set -e

go get github.com/jteeuwen/go-bindata/...

echo "Generating i18n resource file ..."
$GOPATH/bin/go-bindata -pkg resources -o resources/i18n\_resources.go i18n/resources