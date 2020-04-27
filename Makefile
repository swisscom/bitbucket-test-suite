.DEFAULT_GOAL := help

install-dependencies: ## install the go dependencies
	GO111MODULE=on go get github.com/DATA-DOG/godog/cmd/godog@v0.9.0
	go get -u gopkg.in/src-d/go-git.v4/...

execute-test-suite: ## execute the bitbucket testsuite
	go test

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

