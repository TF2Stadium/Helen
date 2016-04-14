#!/usr/bin/env bash

pkg=github.com/TF2Stadium/Helen/internal/version
git_commit=$(git rev-parse HEAD)
git_branch=$(git rev-parse --abbrev-ref HEAD)
build_date=$(date)
build_hostname=$(hostname)

function download_assets {
    echo "Downloading assets"
    curl -o "assets/geoip.mmdb.gz" \
	 "http://geolite.maxmind.com/download/geoip/database/GeoLite2-Country.mmdb.gz"
    gzip -d -f assets/geoip.mmdb.gz
}

function make_assets {
    if [ "$1" = "production" ]; then
	echo "Compiling assets"
	go-bindata -nomemcopy -ignore="bindata\.go"     \
		   -pkg assets 			        \
		   -o assets/bindata.go assets/
    else
	echo "Compiling assets (debug)"
	go-bindata -debug -ignore="bindata\.go"         \
		   -pkg assets    		        \
		   -o assets/bindata.go assets/

    fi
}

function make_binary {
    if [ "$1" = "production" ]; then
	export CGO_ENABLED=0
	echo "Creating production build"

	go build -tags "netgo" -v -ldflags           \
	   "-X \"${pkg}.GitCommit=${git_commit}\"    \
   	   -X \"${pkg}.GitBranch=${git_branch}\"     \
   	   -X \"${pkg}.BuildDate=${build_date}\"     \
   	   -X \"${pkg}.Hostname=${build_hostname}\"" \
	   -o Helen
    else
	echo "Creating development build"

	go build -v -ldflags                         \
	   "-X \"${pkg}.GitCommit=${git_commit}\"    \
   	   -X \"${pkg}.GitBranch=${git_branch}\"     \
   	   -X \"${pkg}.BuildDate=${build_date}\"     \
   	   -X \"${pkg}.Hostname=${build_hostname}\"" \
	   -o Helen
    fi
}

if [ ! -f assets/geoip.mmdb ]; then
    download_assets
fi

make_assets $1
make_binary $1
