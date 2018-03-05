.DEFAULT_GOAL := help

install-dependencies: ## install the go dependencies
	@go get github.com/DATA-DOG/godog/cmd/godog
	@go get -u gopkg.in/src-d/go-git.v4/...
	@go get github.com/satori/go.uuid

execute-test-suite: ## execute the bitbucket testsuite
	@go test

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

