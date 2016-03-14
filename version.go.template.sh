#!/bin/sh

cat <<EOF
package main

import "fmt"

type VersionInfo struct {
	appname   string
	version   string
	buildDate string
	buildHost string
	buildUser string
}

func (v VersionInfo) Block() string {
	return fmt.Sprintf(
		"\n%s Version %s\n\n\tbuild date:\t%s\n\tbuild host:\t%s\n\tbuild user:\t%s\n",
		v.appname,
		v.version,
		v.buildDate,
		v.buildHost,
		v.buildUser,
	)
}

func (v VersionInfo) OneLine() string {
	return fmt.Sprintf(
		"%s Version %s built %s on %s by %s",
		v.appname,
		v.version,
		v.buildDate,
		v.buildHost,
		v.buildUser,
	)
}

var versionInfo VersionInfo = VersionInfo{
	"PullDeploy",
	"$1",
	"`date`",
	"`uname -n`",
	"$USER",
}
EOF
