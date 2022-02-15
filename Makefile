SHELL := /bin/bash
CURRENT_PATH = $(shell pwd)

GO  = GO111MODULE=on go

help: Makefile
	@echo "Choose a command run:"
	@sed -n 's/^##//p' $< | column -t -s ':' | sed -e 's/^/ /'

prepare:
	cd scripts && bash prepare.sh

## make test-coverage: Test project with cover
test-coverage:
	@go test -short -coverprofile cover.out -covermode=atomic ${TEST_PKGS}
	@cat cover.out >> coverage.txt

## make fabric1.4: build fabric(1.4) client plugin
fabric2.3:
	mkdir -p build
# -gcflags="all=-N -l" -trimpath
	$(GO) build --buildmode=plugin -o build/fabric2.3.so ./*.go

docker:
	mkdir -p build
	cd build && rm -rf pier && cp -r ../../pier pier
	cd ${CURRENT_PATH}
	docker build -t meshplus/pier-fabric .

fabric1.4-linux:
	cd scripts && sh cross_compile.sh linux-amd64 ${CURRENT_PATH}

## make linter: Run golanci-lint
linter:
	golangci-lint run
