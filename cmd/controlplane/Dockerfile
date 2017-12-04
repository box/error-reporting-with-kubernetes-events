
# Start from a Debian image with Go 1.9.2 installed
# and a workspace (GOPATH) configured at /go.
FROM golang:1.9.2

# TODO: It looks like docker has a bug while COPY'ing a large number
# of files from host to container during build. Instead of COPYing all files,
# create a tar and then ADD the tar to the destination directory.
# In order to create the tarfile execute `tar cvf allfiles.tar .` from the
# project root.
ADD allfiles.tar /go/src/github.com/box/error-reporting-with-kubernetes-events/
# COPY / /go/src/github.com/box/error-reporting-with-kubernetes-events/

# Build the controlplaneapp command inside the container.
RUN go install github.com/box/error-reporting-with-kubernetes-events/cmd/controlplane

# Run the controlplaneapp is in the PATH located in /go/bin/controlplane.

