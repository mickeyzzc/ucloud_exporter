#!/usr/bin/env bash

prog=ucloud_exporter
version=0.0.1

# 交叉编译
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "\
-X main.version=$version \
-X main.gitCommit=`git rev-parse HEAD` \
-X main.buildTime=`date -u '+%Y-%m-%d_%H:%M:%S'` \
" -v -o $prog main.go
