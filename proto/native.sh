#!/bin/bash

curDir=$(pwd)

if [ "$#" -eq 1 ]; then
    protoFile=$1
elif [ "$#" -eq 2 ]; then
    protoFile="./$1/$2.proto"
else
    protoFile=""
fi

GOMODCACHE=$(go env GOMODCACHE)

if [ "$protoFile" = "" ] || [ ! -f "${protoFile}" ]; then
    echo "Usage: $0 <dir> <proto_file>"
    exit 1
fi

protoc -I=. -I=../ -I=../../ -I="$GOMODCACHE" -I="$GOMODCACHE/github.com/gogo/protobuf/protobuf" --gofast_out=paths=source_relative:. "$protoFile"
