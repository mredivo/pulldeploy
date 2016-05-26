# PullDeploy

Pulldeploy manages a repository of single-artifact application deployment packages, and
deploys and releases those packages on their target hosts.
 
## Documentation

* Project Wiki: https://github.com/mredivo/pulldeploy/wiki
* GoDoc: [![GoDoc](https://godoc.org/github.com/mredivo/pulldeploy?status.svg)](https://godoc.org/github.com/mredivo/pulldeploy)

## Features

* After daemon is installed, no further interaction with application hosts is required
* Artifact repository management can be automated in CI, or done from laptops
* Artifact repository is normally in S3, but can reside elsewhere
* Artifacts can be signed, and will not be deployed if HMAC checking fails
* Versions can have arbitrary names; VCS SHA1, CI build number, etc.
* Deployment can be to a subset of the hosts running an application
* Multiple applications can be managed on one application host
* No commands or the like from the artifact repository are trusted, other than the application itself

## Terminology

"I'm going to deploy a new version of the Foo application to the stage environment,
and release it at 2:00pm. The artifact is already in the repository."

## Developer Setup

### Prerequisites

* A working Go development environment rooted at $GOPATH
* AWS credentials for access to S3, if using S3 for repository storage

### Getting the Source

In addition to the project source, you will also need any dependencies, such
as the AWS libraries. There is also some user-specific configuration that
must be generated. The following steps take care of all this:

```
mkdir -p $GOPATH/src/github.com/mredivo
cd $GOPATH/src/github.com/mredivo
git clone git@github.com:mredivo/pulldeploy.git
cd pulldeploy
make fetch
make devenv
```

### Building and Running

To build the application:

```
make
```

This builds the application into the ./build directory, from where it can be executed:

```
./build/pulldeploy daemon -env=stage
```

* See the Wiki for notes on installing PullDeploy in a production environment.
* See the GoDocs for a command line synopsis, or execute `./build/pulldeploy help`
