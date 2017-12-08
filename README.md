# error-reporting-with-kubernetes-events

Example containers running on kubernetes utilizing kubernetes events.

There are two families of containers, from control plane and application level.
The application level containers, can trigger additional processing in the
control plane container. CustomResourceDefinitions are used for triggering the
processing and passing parameters to the computation. The control plane container
does some checking of the passed parameters. If there is any unexpected values
in the parameters, it uses *Kubernetes Events* to relay the error information
back to the application container.


# Prerequisites

Even though the code *should* work in later versions, this example is tested
with kubernetes 1.7.5. More specifically:

```
$ kubectl version
Client Version: version.Info{Major:"1", Minor:"7", GitVersion:"v1.7.5", GitCommit:"17d7182a7ccbb167074be7a87f0a68bd00d58d97", GitTreeState:"clean", BuildDate:"2017-08-31T19:32:12Z", GoVersion:"go1.9", Compiler:"gc", Platform:"darwin/amd64"}
Server Version: version.Info{Major:"1", Minor:"7", GitVersion:"v1.7.5", GitCommit:"17d7182a7ccbb167074be7a87f0a68bd00d58d97", GitTreeState:"clean", BuildDate:"2017-10-04T09:07:46Z", GoVersion:"go1.8.3", Compiler:"gc", Platform:"linux/amd64"}
```

> NOTE: Because of the lack of availability of
> [CustomResourceDefinitions](https://kubernetes.io/docs/tasks/access-kubernetes-api/extend-api-custom-resource-definitions/),
> this example will not work with kubernetes versions earlier than 1.7.


On a MacOs using
[minikube](https://kubernetes.io/docs/getting-started-guides/minikube/#specifying-the-kubernetes-version) you can
start a kubernetes server at v1.7.5.
```
minikube start --kubernetes-version v1.7.5
```
[brew](https://brew.sh/) can be used to install an old kubectl at
[v1.7.5](https://github.com/Homebrew/homebrew-core/blob/8e0e4c9c9b1c4154f31f3313e6b5cfce7de79109/Formula/kubernetes-cli.rb#L5).

You should use the docker daemon in the minikube vm. Follow the instructions
[here](https://kubernetes.io/docs/getting-started-guides/minikube/#reusing-the-docker-daemon)

Make sure you are using a recent version of docker. This is tested using these
versions

```
$ docker version
Client:
 Version:      17.11.0-ce
 API version:  1.23
 Go version:   go1.9.2
 Git commit:   1caf76c
 Built:        unknown-buildtime
 OS/Arch:      darwin/amd64

Server:
 Version:      17.06.0-ce
 API version:  1.30 (minimum version 1.12)
 Go version:   go1.8.3
 Git commit:   02c1d87
 Built:        Fri Jun 23 21:51:55 2017
 OS/Arch:      linux/amd64
 Experimental: false
```

> NOTE: For older versions of docker, we have seen problems with the `COPY`
> command in Dockerfiles while copying a large number of files.

# Build

## Build the docker containers

From project root execute:

```
docker build  -f cmd/controlplane/Dockerfile \
   -t boxinc/error-reporting-with-kubernetes-events:controlplane  .
```

# Run

## Start the control plane app and define the CustomResourceDefinition

from repo root
```
kubectl  apply  -f cmd/controlplane/config.yml
```

Expected output:
```
customresourcedefinition "pkis.box.com" configured
namespace "controlplane" configured
deployment "controlplane" configured
```

## Start the applications that trigger processing in the control plane app





