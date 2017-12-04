
# Start from a Debian image with Go 1.9.2 installed
# and a workspace (GOPATH) configured at /go.
FROM golang:1.9.2

# Copy the local package files to the container's workspace.
ADD . /go/src/github.com/box/error-reporting-with-kubernetes-events
RUN ls -all /go/src/github.com/box/error-reporting-with-kubernetes-events
RUN ls -all /go/src/github.com/box/error-reporting-with-kubernetes-events/cmd
RUN ls -all /go/src/github.com/box/error-reporting-with-kubernetes-events/cmd/controlplane

# Build the controlplaneapp command inside the container.
RUN go install github.com/box/error-reporting-with-kubernetes-events/cmd/controlplane

# Run the controlplaneapp is located in /go/bin/controlplane

