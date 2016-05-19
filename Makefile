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

devenv: $(VERSIONINFO) data/etc/$(TARGET).yaml

fetch:
	go get -t -d -v ./...

test: $(VERSIONINFO) data/etc/$(TARGET).yaml
	go test -race $(TESTS)

$(VERSIONINFO): VERSION make_versioninfo.sh $(SOURCES)
	@echo "Generating $(VERSIONINFO) with version \"$(VERSION)\""
	$(shell ./make_versioninfo.sh $(VERSION) > $(VERSIONINFO))

data/etc/$(TARGET).yaml: yaml/prototype.yaml
ifdef CIRCLECI
	sed -e "s/USERNAME/circleci/" -e "s#PROJECTDIR#$(PWD)#" < yaml/prototype.yaml > data/etc/$(TARGET).yaml
else
	sed -e "s/USERNAME/$(USER)/" -e "s#PROJECTDIR#$(PWD)#" < yaml/prototype.yaml > data/etc/$(TARGET).yaml
endif

build: $(BUILDDIR)/$(TARGET)

$(BUILDDIR)/$(TARGET): $(SOURCES) doc.go $(VERSIONINFO)
	go build -o $(BUILDDIR)/$(TARGET) $(TARGET).go doc.go
