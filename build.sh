#!/bin/sh
set -e
go test
go build -o mpv-subserv.so -buildmode=c-shared entf.net/mpv-subserv
