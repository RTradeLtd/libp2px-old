.PHONY: proto-gen
proto-gen:
	protoc \
		-I=p2p/protocol/identify \
		-I=${GOPATH}/src \
		-I=${GOPATH}/src/github.com/gogo/protobuf/protobuf \
		p2p/protocol/identify/pb/identify.proto \
		--gogofaster_out=p2p/protocol/identify

# run standard go tooling for better readability
.PHONY: tidy
tidy: imports fmt
	go vet ./...
	golint ./...

# automatically add missing imports
.PHONY: imports
imports:
	find . -type f -name '*.go' -exec goimports -w {} \;

# format code and simplify if possible
.PHONY: fmt
fmt:
	find . -type f -name '*.go' -exec gofmt -s -w {} \;