# Build the manager binary
FROM golang:1.13 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o manager main.go

# build oci-image-tool
WORKDIR /workspace-oci-image-tool
RUN go get -d github.com/opencontainers/image-tools/cmd/oci-image-tool
RUN make -C /go/src/github.com/opencontainers/image-tools/ tool
RUN cp /go/src/github.com/opencontainers/image-tools/oci-image-tool /go/bin/oci-image-tool

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
#FROM gcr.io/distroless/static:nonroot
#FROM fedora:31
#RUN dnf install -y skopeo net-tools bind-utils iputils
FROM alpine:latest
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
RUN apk add skopeo
WORKDIR /
COPY --from=builder /workspace/manager .
COPY --from=builder /go/src/github.com/opencontainers/image-tools/oci-image-tool /usr/bin/oci-image-tool
#USER nonroot:nonroot

ENTRYPOINT ["/manager"]
