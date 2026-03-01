BINARY_NAME=terraform-provider-atlassian
GOPATH=$(shell go env GOPATH)
GOLANGCI_LINT_VERSION=v1.64.5

default: build

.PHONY: build
build:
	go build -o $(BINARY_NAME)

.PHONY: install
install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/atlassian/atlassian/0.1.0/linux_amd64
	cp $(BINARY_NAME) ~/.terraform.d/plugins/registry.terraform.io/atlassian/atlassian/0.1.0/linux_amd64/

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: fmt
fmt:
	gofmt -s -w .

.PHONY: test
test:
	go test -race -count=1 ./...

.PHONY: testacc
testacc:
	TF_ACC=1 go test -v -count=1 -timeout 30m ./...

.PHONY: docs
docs:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate

.PHONY: clean
clean:
	rm -f $(BINARY_NAME)
