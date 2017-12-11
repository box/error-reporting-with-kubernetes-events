
# Start from a Debian image with Go 1.9.2 installed
# and a workspace (GOPATH) configured at /go.
FROM golang:1.9.2

# TODO: This COPY command copies a large number of files. We have seen this to
# fail in older docker versions like 1.10.X
COPY / /go/src/github.com/box/error-reporting-with-kubernetes-events/

# An input file to the controlplane application.
COPY /cmd/controlplane/AllowedNames /

# Build the controlplaneapp command inside the container.
RUN go install github.com/box/error-reporting-with-kubernetes-events/cmd/controlplane

# Run the controlplaneapp is in the PATH located in /go/bin/controlplane.

