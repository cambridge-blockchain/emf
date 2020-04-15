SHELL = /bin/bash

.PHONY: help run install clean test patch deps

help: ## Shows this help
	@IFS=$$'\n' ; \
	help_lines=(`fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##/:/'`); \
	printf "%-30s %s\n" "command" "description" ; \
	printf "%-30s %s\n" "-------" "-----------" ; \
	for help_line in $${help_lines[@]}; do \
		IFS=$$':' ; \
		help_split=($$help_line) ; \
		help_command=`echo $${help_split[0]} | sed -e 's/^ *//' -e 's/ *$$//'` ; \
		help_info=`echo $${help_split[2]} | sed -e 's/^ *//' -e 's/ *$$//'` ; \
		printf '\033[36m'; \
		printf "%-30s %s" $$help_command ; \
		printf '\033[0m'; \
		printf "%s\n" $$help_info; \
	done

clean: ## Removes all dependencies that can be auto-reinstalled via glide
	go clean
	rm -rf ./vendor/

patch: ## Updates dependencies if glide.yaml has changed
	GO111MODULE=on go get -u=patch
	GO111MODULE=on go mod tidy

deps:
	GO111MODULE=on go build -v ./...

vendor/: ## Install all golang dependencies via glide, useful for multi-arch docker builds
	GO111MODULE=on go mod vendor

install: clean vendor/ ## Install all golang dependencies via glide, useful for multi-arch docker builds

run: deps ## Runs the component natively on the local machine
	CONFIG=./config.yaml go run -gcflags "all=-trimpath=$$GOPATH" example-main.go

test: ## Runs the go unit testing system
	go test -v -timeout 60000ms -cover `go list ./... | grep -v /vendor/`

lint: ## Runs gometalinter against the local components
	golangci-lint run
