export IPFS_API ?= v04x.ipfs.io

gx:
	go get -u github.com/whyrusleeping/gx
	go get -u github.com/whyrusleeping/gx-go

deps: gx
	gx --verbose install --global
	gx-go rewrite

build: deps
	export PATH=${PATH}:${TRAVIS_HOME}/gopath/src/github.com/Ice-Storm/cherrychain
	export GOPATH=${PATH}:${TRAVIS_HOME}/gopath/src/github.com/Ice-Storm/cherrychain
	go build main.go

publish:
	gx-go rewrite --undo
