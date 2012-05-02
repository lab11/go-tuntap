#!/bin/bash

GOOS=`go env GOOS`

case $GOOS in
    linux)
        go tool cgo -godefs=true types_linux.go >ztypes_linux.go
        rm -rf _obj
        ;;
    *)
        echo "Don't know how to compile types for $GOOS"
        exit 1
esac
