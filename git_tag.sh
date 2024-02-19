#!/bin/sh

[[ "$#" -ne  "2" ]] && echo "Provide version then message" && exit 1

go generate
git add ./version.go
git commit -m $2
git tag -a $1 -m $2
