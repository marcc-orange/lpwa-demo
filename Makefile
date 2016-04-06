GO_SRC=$(shell find . -type f -name '*.go')
STATIC_FILES=$(wildcard static/*)
BUILD_STATIC_FILES=$(patsubst static/%, build/static/%, $(STATIC_FILES))
SERVER=http://localhost:8080


# Deploy using a git push
deploy: build/demo $(BUILD_STATIC_FILES)
	cd build && git commit -a --amend -m "deploy app" && git push -f origin master

# Build for linux, move to build/
build/demo: $(GO_SRC)
	mkdir -p build
	env GOOS=linux GOARCH=386 go build
	mv demo build/demo

# Copy static resources to build/
build/static/%: static/%
	mkdir -p build
	cp -f $< $@

# Simulate a push from datavenue
push:
	curl -X POST -d @example/sample.json $(SERVER)/push
