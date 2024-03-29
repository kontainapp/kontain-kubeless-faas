#!/bin/bash

if [ -z "$GOPATH" ]
then
    export GOPATH=$HOME/go
fi

rm -f go.mod go.sum

go get -d github.com/opencontainers/image-tools/cmd/oci-image-tool
make -C $GOPATH/src/github.com/opencontainers/image-tools tool

# The oci-image-tool executable should be $GOPATH/src/github.com/opencontainers/image-tools/oci-image-tool
cp $GOPATH/src/github.com/opencontainers/image-tools/oci-image-tool .
