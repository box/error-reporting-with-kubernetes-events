
# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang

# Copy the local package files to the container's workspace.
ADD . /go/src/github.com/Box/controlplaneapp

# Build the controlplaneapp command inside the container.
RUN go install github.com/Box/controlplaneapp

# Run the controlplaneapp command by default when the container starts.
ENTRYPOINT /go/bin/controlplaneapp

