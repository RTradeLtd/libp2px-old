# Contributing

While we won't be overtly strict about accepting outside commits to the codebase, however to ensure that the codebase stays healthy, we expect the following:

* All new changes must be covered by *quality* unit tests
* All changes must pass the following CI checks:
  * `staticcheck`
  * `gofmt`
  * `golint`
  * `go vet`
* Any scripts must pass https://www.shellcheck.net/
