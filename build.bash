#!/usr/bin/env bash

pkg=github.com/TF2Stadium/Helen/internal/version
git_commit=$(git rev-parse HEAD)
git_branch=$(git rev-parse --abbrev-ref HEAD)
build_date=$(date)
build_hostname=$(hostname)

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

bash build_assets.bash $1
make_binary $1
