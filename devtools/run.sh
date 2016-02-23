#!/bin/sh

go run -race pulldeploy.go cmdrepo.go cmdrelease.go cmdinfo.go $@
