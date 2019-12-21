###############################################
#
# Makefile
#
###############################################

.DEFAULT_GOAL := build

.PHONY: test

GOPATH = "${PWD}"

lint:
	GOPATH=${GOPATH} ~/go/bin/golint dom.go

deps:
	GOPATH=${GOPATH} go get -d golang.org/x/net/html

build: deps
	GOPATH=${GOPATH} go build .

test: build
	GOPATH=${GOPATH} go test -v .