# error-reporting-with-kubernetes-events

Example apps running on kubernetes utilizing kubernetes events.


# Installation

Even though the code *should* work in later versions, this example is tested
with kubernetes 1.7.5. More specifically:

```
$ kubectl version
Client Version: version.Info{Major:"1", Minor:"7", GitVersion:"v1.7.5", GitCommit:"17d7182a7ccbb167074be7a87f0a68bd00d58d97", GitTreeState:"clean", BuildDate:"2017-08-31T19:32:12Z", GoVersion:"go1.9", Compiler:"gc", Platform:"darwin/amd64"}
Server Version: version.Info{Major:"1", Minor:"7", GitVersion:"v1.7.5", GitCommit:"17d7182a7ccbb167074be7a87f0a68bd00d58d97", GitTreeState:"clean", BuildDate:"2017-10-04T09:07:46Z", GoVersion:"go1.8.3", Compiler:"gc", Platform:"linux/amd64"}
```

On a MacOs using
[minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/) you can
start a kubernetes server at v1.7.5.
[brew](https://brew.sh/) can be used to install an old kubectl at
[v1.7.5](https://github.com/Homebrew/homebrew-core/blob/8e0e4c9c9b1c4154f31f3313e6b5cfce7de79109/Formula/kubernetes-cli.rb#L5).


