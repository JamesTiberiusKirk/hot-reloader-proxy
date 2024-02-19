#!/bin/sh

version=`git describe --tags`

if [ "$1" ]; then
    version=$version:$1
fi

cat << EOF > version.go
package hotreloaderproxy

//go:generate sh ./get_version.sh
var Version = "$version"
EOF

# Inspired/copied from: https://adrianhesketh.com/2016/09/04/adding-a-version-number-to-go-packages-with-go-generate/
