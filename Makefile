default: install

.PHONY: build
build:
	go build -o terraform-provider-outpost

.PHONY: install
install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/outpost/outpost/0.1.0/$$(go env GOOS)_$$(go env GOARCH)
	cp terraform-provider-outpost ~/.terraform.d/plugins/registry.terraform.io/outpost/outpost/0.1.0/$$(go env GOOS)_$$(go env GOARCH)/

.PHONY: test
test:
	go test -v -cover ./...

.PHONY: testacc
testacc:
	TF_ACC=1 go test -v -cover ./...

.PHONY: fmt
fmt:
	go fmt ./...
	terraform fmt -recursive ./examples/

.PHONY: lint
lint:
	golangci-lint run

.PHONY: generate
generate:
	go generate ./...

.PHONY: clean
clean:
	rm -f terraform-provider-outpost
	rm -rf dist/

.PHONY: docs
docs:
	go generate ./...

