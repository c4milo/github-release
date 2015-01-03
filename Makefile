VERSION := v1.0.0

compile: deps
	go build

install:
	go install

deps: 
	go get github.com/c4milo/github-release

release: 
	set -x; \
	latest_tag=$$(git describe --tags `git rev-list --tags --max-count=1`); \
	comparison="$$latest_tag..HEAD"; \
	if [ -z "$$latest_tag" ]; then comparison=""; fi; \
	changelog=$$(git log $$comparison --oneline --no-merges --reverse); \
	DEBUG=1 github-release c4milo/github-release $(VERSION) "$$(git rev-parse --abbrev-ref HEAD)" "**Changelog**<br/>$$changelog" ${FILES}; \
	git pull 
