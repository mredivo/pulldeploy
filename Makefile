# Makefile for PullDeploy

TARGET = pulldeploy

BUILDDIR = build

SRCLIBS = ./configloader/*go ./deployment/*go ./repostorage/*go ./signaller/*go
SOURCES = $(TARGET).go $(SRCLIBS)
TESTS = ./configloader ./deployment ./repostorage ./signaller
PWD = $(shell pwd)

ifdef CIRCLECI
	VERSION = $(shell cat VERSION).$(CIRCLE_BUILD_NUM)
else
	VERSION = $(shell cat VERSION).0
endif
SHA = $(shell git rev-parse --short HEAD)
ARTIFACT = $(TARGET)-$(VERSION).tar.gz

all: build

clean:
	rm -rf $(BUILDDIR)
	rm -rf data/client/*
	rm -rf data/repository/*

devclean: clean
	rm -f version.go
	rm -f data/etc/$(TARGET).yaml
	rm -rf data/etc/pulldeploy.d/*

devenv: version.go data/etc/$(TARGET).yaml

fetch:
	go get -t -d -v ./...

test: version.go data/etc/$(TARGET).yaml
	go test -race $(TESTS)

version.go: VERSION version.go.template.sh $(SOURCES)
	@echo "Generating version.go with version \"$(VERSION)\""
	$(shell ./version.go.template.sh $(VERSION) > version.go)

data/etc/$(TARGET).yaml: yaml/prototype.yaml
ifdef CIRCLECI
	sed -e "s/USERNAME/circleci/" -e "s#PROJECTDIR#$(PWD)#" < yaml/prototype.yaml > data/etc/$(TARGET).yaml
else
	sed -e "s/USERNAME/$(USER)/" -e "s#PROJECTDIR#$(PWD)#" < yaml/prototype.yaml > data/etc/$(TARGET).yaml
endif

build: $(BUILDDIR)/$(TARGET)

$(BUILDDIR)/$(TARGET): $(SOURCES)
	go build -o $(BUILDDIR)/$(TARGET) $(TARGET).go version.go
