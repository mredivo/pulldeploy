# PullDeploy

Pulldeploy manages a repository of single-artifact application deployment packages, and
deploys and releases those packages on their target hosts.
 
## Documentation

* Project Wiki: https://github.com/mredivo/pulldeploy/wiki
* GoDoc: [![GoDoc](https://godoc.org/github.com/mredivo/pulldeploy?status.svg)](https://godoc.org/github.com/mredivo/pulldeploy)

## Features

*General*

* Application servers fetch the app and release it automatically and unattended; no server enumeration required
* Releases can be to a subset of the hosts running an application
* Rollback is as easy as re-releasing a previous version

*Management*

* All management is done through a command line utility
* Repository management can be automated in CI, or done manually from laptops
* Artifacts are stored in Amazon S3 (with provision for alternate storage)

*Configurability*

* Versions can have arbitrary names; VCS SHA or revision, CI build number, etc.
* Custom artifact types can be defined, along with the command to unpack them
* Multiple applications can be managed on one application host

*Security*

* Artifacts are signed, and will not be deployed if HMAC checking fails
* Ownership of all deployed files is set to specified (non-root) user
* No commands from the artifact repository are trusted, other than the application itself
* Command line utilities do not require root privileges
* When run as root, daemon will not execute commands from insecure configuration files
* Daemon can be run as non-root (provided the client app can be restarted as non-root)

## Terminology

"I'm going to deploy a new version of the Foo application to the staging environment,
and release it at 2:00pm. The artifact has already been uploaded to the repository."

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
./build/pulldeploy daemon -env=staging
```

* See the Wiki for notes on installing PullDeploy in a production environment.
* See the GoDocs for a command line synopsis, or execute `./build/pulldeploy help`

## Credits

PullDeploy owes its inspiration to a number of different pull-deploy tools I used
in my time at Zynga. It ended up quite different in implementation, but does the
same basic job.

PullDeploy owes its existence to the generosity and support of Change.org.
