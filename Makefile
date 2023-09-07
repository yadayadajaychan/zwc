.PHONY: build test clean

export CGO_ENABLED = 0
export GOFLAGS = -trimpath

build:
	go build -o zwc main/main.go

test:
	@go test -cover
	@test/test.sh

clean:
	-rm zwc
