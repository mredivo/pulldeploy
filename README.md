# PullDeploy

Pulldeploy manages a repository of single-artifact application deployment packages, and
deploys and releases those packages on their target hosts.
 
## Status

Nearly complete, but not operational yet.

## Features

* Once daemon is installed, no further interaction with application hosts is required
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

## Synopsis

```
$ pulldeploy help

usage: pulldeploy <command> [<args>]

Commands:

    Help:
        -h, help [<command>]
        -v, version

    Repository management:
        pulldeploy initrepo -app=<app>
        pulldeploy addenv   -app=<app> envname [envname envname ...]
        pulldeploy rmenv    -app=<app> envname [envname envname ...]
        pulldeploy set      -app=<app> -env=<env> [-keep=n]

    Release management:
        pulldeploy upload  -app=<app> -version=<version> [-disabled] <file>
        pulldeploy enable  -app=<app> -version=<version>
        pulldeploy disable -app=<app> -version=<version>
        pulldeploy purge   -app=<app> -version=<version>
        pulldeploy deploy  -app=<app> -version=<version> -env=<env>
        pulldeploy release -app=<app> -version=<version> -env=<env> [host1, host2, ...]

    Informational:
        pulldeploy list
        pulldeploy status -app=<app>
        pulldeploy listhosts -app=<app> -env=<env>

    Daemon:
        pulldeploy daemon -env=<env> <daemon args>...

```
