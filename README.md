# a test suite for bitbucket

## prerequisites
* install go
* `go get github.com/DATA-DOG/godog/cmd/godog`
* `go get -u gopkg.in/src-d/go-git.v4/...`
* set env `BB_HTTP_URL`
* set env `BB_SSH_URL`
* set env `BB_PROJECT`
* set env `BB_USER`
* set env `BB_PASSWORD`
* set env `BB_REPOSITORY`

## test suite execution
* `go test`
