export IPFS_API ?= v04x.ipfs.io

gx:
	go get -u github.com/whyrusleeping/gx
	go get -u github.com/whyrusleeping/gx-go

deps: gx
	gx --verbose install --global
	gx-go rewrite

build: deps
	# go build main.go

publish:
	gx-go rewrite --undo
