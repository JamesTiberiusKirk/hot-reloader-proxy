#!/bin/sh

[[ "$#" -ne  "2" ]] && echo "Provide version then message" && exit 1

go generate
git tag -a $1 -m $2
