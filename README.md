# error-reporting-with-kubernetes-events

[![Project Status](http://opensource.box.com/badges/stable.svg)](http://opensource.box.com/badges)
[![CircleCI](https://circleci.com/gh/box/error-reporting-with-kubernetes-events.svg?style=svg)](https://circleci.com/gh/box/error-reporting-with-kubernetes-events)

Example containers running on kubernetes utilizing kubernetes events.

There are two families of containers, from control plane and application level.
The application level containers, can trigger additional processing in the
control plane container. CustomResourceDefinitions are used for triggering the
processing and passing parameters to the computation. The control plane
container does some checking of the passed parameters. If there is any
unexpected values in the parameters, [*Kubernetes
Events*](https://v1-7.docs.kubernetes.io/docs/api-reference/v1.7/#event-v1-core)
are used to relay the error information back to the application container.


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

From repo root:
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


From repo root:
```
kubectl apply  -f cmd/app1/config.yml
kubectl apply  -f cmd/app2/config.yml
```

Expected output:

```
namespace "app1" configured
pki "app1-pki" configured
deployment "app1" configured

namespace "app2" configured
pki "app2-pki" configured
deployment "app2" created
```


# Expected output

All [pods](https://kubernetes.io/docs/concepts/workloads/pods/pod/) except the
pod for `app1` should transition to `Running` state eventually.

```
$ kubectl get pods --all-namespaces
NAMESPACE      NAME                            READY     STATUS              RESTARTS   AGE
app1           app1-647877271-lrpbd            0/1       ContainerCreating   0          1d
app2           app2-3513338968-9hf6l           1/1       Running             1          1d
controlplane   controlplane-1148696967-13dgv   1/1       Running             0          1d
....
```

There is an error in `app1`'s `config.yml` to demonstrate the error
handling using [Kubernetes
Events](https://v1-7.docs.kubernetes.io/docs/api-reference/v1.7/#event-v1-core)

The events history in `kubectl describe` output provides diagnostic
information to the application owner so that she can fix the error
easily.

```
kubectl describe pod app1-647877271-lrpbd --namespace app1

Events:
  FirstSeen   LastSeen   Count   From         SubObjectPath   Type      Reason         Message
  ---------   --------   -----   ----         -------------   --------   ------         -------
  ....
  1d      1m      24   controlplane            Warning      pki ServiceName error   ServiceName: appp1 in pki: app1-pki is not found in allowedNames: [app1 app2]
  ....

```

The shown error line is generated by the controlplane app and placed at the
event stream of the problematic pod. More details about generating this
error event can be found at the controlplane application's implementation.

The application programmer can easily rootcause that the `serviceName` field
in `config.yml` has a typo.
```
spec:
  # This service name has a typo
  serviceName: appp1
```





## Support

Need to contact us directly? Email oss@box.com and be sure to include the name of this project in the subject.

## Copyright and License

Copyright 2017 Box, Inc. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
