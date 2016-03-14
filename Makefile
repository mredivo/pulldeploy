# Makefile for PullDeploy

TARGET = pulldeploy

BUILDDIR = build

SRCMAIN = cmdinfo.go cmdrepo.go cmdrelease.go
SRCLIBS = ./configloader/*go ./deployment/*go ./repostorage/*go ./signaller/*go
SOURCES = $(TARGET).go $(SRCMAIN) $(SRCLIBS)
TESTS = ./configloader ./deployment ./repostorage ./signaller

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
	rm -f yaml/$(TARGET).yaml

devenv: version.go yaml/$(TARGET).yaml

fetch:
	go get -t -d -v ./...

test: version.go yaml/$(TARGET).yaml
	go test -race $(TESTS)

version.go: VERSION version.go.template.sh $(SOURCES)
	@echo "Generating version.go with version \"$(VERSION)\""
	$(shell ./version.go.template.sh $(VERSION) > version.go)

yaml/$(TARGET).yaml: yaml/prototype.yaml
ifdef CIRCLECI
	sed -e "s/USERNAME/circleci/" -e "s#~#$(HOME)#" < yaml/prototype.yaml > yaml/$(TARGET).yaml
else
	sed -e "s/USERNAME/$(USER)/" -e "s#~#$(HOME)#" < yaml/prototype.yaml > yaml/$(TARGET).yaml
endif

# "make build" (the default) to build a local native binary on any machine.
build: $(BUILDDIR)/$(TARGET)

$(BUILDDIR)/$(TARGET): $(SOURCES)
	go build -o $(BUILDDIR)/$(TARGET) $(TARGET).go $(SRCMAIN) version.go

# "make package" to cross-compile a Linux binary on your Mac.
package: $(BUILDDIR)/linux-amd64/$(TARGET)

$(BUILDDIR)/linux-amd64/$(TARGET): $(SOURCES)
	mkdir -p $(BUILDDIR)/linux-amd64
	cp yaml/$(TARGET).yaml $(BUILDDIR)/linux-amd64/$(TARGET).yaml
	env GOOS=linux GOARCH=amd64 go build -o $(BUILDDIR)/linux-amd64/$(TARGET) $(TARGET).go $(SRCMAIN) version.go
	tar czf $(BUILDDIR)/$(ARTIFACT) -C $(BUILDDIR)/linux-amd64 .
	@echo "\npackaged Linux version $(VERSION) into $(BUILDDIR)/$(ARTIFACT)\n"

#upload: $(BUILDDIR)/linux-amd64/$(TARGET)
#	aws s3 cp $(BUILDDIR)/$(ARTIFACT) s3://change-deployable-builds/$(TARGET)/
