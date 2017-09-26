SHELL := /bin/bash

REV := $(shell git rev-parse HEAD)
BRANCH := $(shell git symbolic-ref HEAD 2>/dev/null | cut -d'/' -f 3)
CHANGES := $(shell test -n "$$(git status --porcelain)" && echo '+CHANGES' || true)

TARGET := ofsrvr
VERSION := $(shell cat VERSION)

DEPLOY_IMAGE := $(TARGET)
DEPLOY_ACCOUNT := overtonestudio

OS := darwin linux windows
ARCH := 386 amd64

.PHONY: \
	help \
	default \
	clean \
	clean-artifacts \
	clean-releases \
	clean-vendor \
	tools \
	deps \
	test \
	coverage \
	vet \
	errors \
	lint \
	imports \
	fmt \
	env \
	build \
	build-all \
	doc \
	release \
	package-release \
	sign-release \
	check \
	vendor \
	version

all: imports fmt lint vet errors build

help:
	@echo 'Usage: make <OPTIONS> ... <TARGETS>'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@echo '    help               Show this help screen.'
	@echo '    clean              Remove binaries, artifacts and releases.'
	@echo '    clean-artifacts    Remove build artifacts only.'
	@echo '    clean-releases     Remove releases only.'
	@echo '    clean-vendor       Remove content of the vendor directory.'
	@echo '    tools              Install tools needed by the project.'
	@echo '    deps               Download and install build time dependencies.'
	@echo '    test               Run unit tests.'
	@echo '    coverage           Report code tests coverage.'
	@echo '    vet                Run go vet.'
	@echo '    errors             Run errcheck.'
	@echo '    lint               Run golint.'
	@echo '    imports            Run goimports.'
	@echo '    fmt                Run go fmt.'
	@echo '    env                Display Go environment.'
	@echo '    build              Build project for current platform.'
	@echo '    build-all          Build project for all supported platforms.'
	@echo '    doc                Start Go documentation server on port 8080.'
	@echo '    release            Package and sing project for release.'
	@echo '    package-release    Package release and compress artifacts.'
	@echo '    check              Verify compiled binary.'
	@echo '    vendor             Update and save project build time dependencies.'
	@echo '    version            Display Go version and App version'
	@echo '    docker_build       Build a binary for use with Docker (Linux-Amd64)'
	@echo '    docker-image       Build a docker image'
	@echo '    docker-release     Build a binary and create a Docker image'
	@echo '    docker-deploy      Tag and Push a Docker image to a Docker repository'
	@echo ''
	@echo 'Targets run by default are: imports, fmt, lint, vet, errors and build.'
	@echo ''

print-%:
	@echo $* = $($*)

clean: clean-artifacts clean-releases
	go clean -i ./...
	rm -vf \
	  $(CURDIR)/coverage.* \
	  $(CURDIR)/$(TARGET)_*

clean-artifacts:
	rm -Rf artifacts/*

clean-releases:
	rm -Rf releases/*

clean-vendor:
	find $(CURDIR)/vendor -type d -print0 2>/dev/null | xargs -0 rm -Rf

clean-all: clean clean-artifacts clean-vendor

tools:
	go get golang.org/x/tools/cmd/goimports
	go get github.com/kisielk/errcheck
	go get github.com/golang/lint/golint
	go get github.com/axw/gocov/gocov
	go get github.com/matm/gocov-html
	go get github.com/tools/godep
	go get github.com/mitchellh/gox

deps:
	godep restore

test: deps
	go test -v ./...

coverage: deps
	gocov test ./... > $(CURDIR)/coverage.out 2>/dev/null
	gocov report $(CURDIR)/coverage.out
	if test -z "$$CI"; then \
	  gocov-html $(CURDIR)/coverage.out > $(CURDIR)/coverage.html; \
	  if which open &>/dev/null; then \
	    open $(CURDIR)/coverage.html; \
	  fi; \
	fi

vet:
	go vet -v ./...

errors:
	errcheck -ignoretests -blank ./...

lint:
	golint ./...

imports:
	goimports -l -w .

fmt:
	go fmt ./...

env:
	@go env

build: deps
	go build -v \
		-tags "$(TAGS)"
	   -ldflags "$(LDFLAGS)" \
	   -o "$(TARGET)" .

build-all: deps
	mkdir -v -p $(CURDIR)/artifacts/$(VERSION)
	gox -verbose \
	    -os "$(OS)" -arch "$(ARCH)" \
			-tags "$(TAGS)"
	    -ldflags "$(LDFLAGS)" \
	    -output "$(CURDIR)/artifacts/$(VERSION)/{{.OS}}_{{.Arch}}/$(TARGET)" .
	cp -v -f \
	   $(CURDIR)/artifacts/$(VERSION)/$$(go env GOOS)_$$(go env GOARCH)/$(TARGET) .

run: build
	./ofsrvr -config ./configs/default.config.yml

doc:
	godoc -http=:8088 -index

release: package-release sign-release

package-release:
	@test -x $(CURDIR)/artifacts/$(VERSION) || exit 1
	mkdir -v -p $(CURDIR)/releases/$(VERSION)
	for release in $$(find $(CURDIR)/artifacts/$(VERSION) -mindepth 1 -maxdepth 1 -type d 2>/dev/null); do \
	  platform=$$(basename $$release); \
	  pushd $$release &>/dev/null; \
	  zip $(CURDIR)/releases/$(VERSION)/$(TARGET)_$${platform}.zip $(TARGET); \
	  popd &>/dev/null; \
	done

check:
	@test -x $(CURDIR)/$(TARGET) || exit 1
	if $(CURDIR)/$(TARGET) --version | grep -qF '$(VERSION)'; then \
	  echo "$(CURDIR)/$(TARGET): OK"; \
	else \
	  exit 1; \
	fi

vendor: deps
	godep save

version:
	@go version
	@echo $(VERSION)

docker-build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags '$(TAGS)' -ldflags "$(LDFLAGS) -s -w $(LDFLAGS)" -o artifacts/$(TARGET)

docker-image:
	docker build -t $(DEPLOY_ACCOUNT)/$(DEPLOY_IMAGE) -f Dockerfile .

docker-release: docker-build docker-image

docker-deploy:
ifeq ($(dockertag),)
	@echo "Usage: make $@ dockertag=<dockertag>"
	@exit 1
endif
	docker tag $(DEPLOY_ACCOUNT)/$(TARGET):latest quay.io/$(DEPLOY_ACCOUNT)/$(TARGET):$(dockertag)
	docker push quay.io/$(DEPLOY_ACCOUNT)/$(TARGET):$(dockertag)

docker-run: docker-release
	docker-compose up
