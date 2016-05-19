# Makefile for PullDeploy

TARGET = pulldeploy

BUILDDIR = build

VERSIONINFO = pdconfig/versioninfo.go
SRCLIBS     = pdconfig/pdconfig.go \
              pdconfig/configloader.go \
              deployment/*go \
              storage/*go \
              repo/*go \
              signaller/*go \
              command/*go
SOURCES     = $(TARGET).go $(SRCLIBS)
TESTS       = ./pdconfig ./deployment ./storage ./repo ./signaller
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
	rm -f $(VERSIONINFO)
	rm -f data/etc/$(TARGET).yaml
	rm -rf data/etc/pulldeploy.d/*

devenv: $(VERSIONINFO) data/etc/$(TARGET).yaml data/etc/pulldeploy.d/sample_app.json

fetch:
	go get -t -d -v ./...

test: $(VERSIONINFO) data/etc/$(TARGET).yaml
	go test -race $(TESTS)

$(VERSIONINFO): VERSION make_versioninfo.sh $(SOURCES)
	@echo "Generating $(VERSIONINFO) with version \"$(VERSION)\""
	$(shell ./make_versioninfo.sh $(VERSION) > $(VERSIONINFO))

data/etc/$(TARGET).yaml: data/configs/$(TARGET).yaml
	sed -e "s#PROJECTDIR#$(PWD)#" < data/configs/$(TARGET).yaml > data/etc/$(TARGET).yaml

data/etc/pulldeploy.d/sample_app.json: data/configs/sample_app.json
	sed -e "s#PROJECTDIR#$(PWD)#" < data/configs/sample_app.json > data/etc/pulldeploy.d/sample_app.json

build: $(BUILDDIR)/$(TARGET)

$(BUILDDIR)/$(TARGET): $(SOURCES) doc.go $(VERSIONINFO)
	go build -o $(BUILDDIR)/$(TARGET) $(TARGET).go doc.go
