# This Source Code Form is subject to the terms of the Mozilla Public
# License, version 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at http://mozilla.org/MPL/2.0/.

NAME := github-release
VERSION := v1.0.9

build:
	go build -ldflags "-X main.Version=$(VERSION)"

compile:
	@rm -rf build/
	@gox -ldflags "-X main.Version=$(VERSION)" \
	-osarch="darwin/amd64" \
	-os="linux" \
	-os="windows" \
	-os="solaris" \
	-output "build/{{.Dir}}_$(VERSION)_{{.OS}}_{{.Arch}}/$(NAME)" \
	./...

install:
	go install -ldflags "-X main.Version=$(VERSION)"

deps:
	go get github.com/c4milo/github-release
	go get github.com/mitchellh/gox

dist: compile
	$(eval FILES := $(shell ls build))
	@rm -rf dist && mkdir dist
	@for f in $(FILES); do \
		(cd $(shell pwd)/build/$$f && tar -cvzf ../../dist/$$f.tar.gz *); \
		(cd $(shell pwd)/dist && shasum -a 512 $$f.tar.gz > $$f.sha512); \
		echo $$f; \
	done

release: dist
	@latest_tag=$$(git describe --tags `git rev-list --tags --max-count=1`); \
	comparison="$$latest_tag..HEAD"; \
	if [ -z "$$latest_tag" ]; then comparison=""; fi; \
	changelog=$$(git log $$comparison --oneline --no-merges); \
	github-release c4milo/$(NAME) $(VERSION) "$$(git rev-parse --abbrev-ref HEAD)" "**Changelog**<br/>$$changelog" 'dist/*'; \
	git pull

.PHONY: build compile install deps dist release
