# Copyright © 2021 Kontain Inc. All rights reserved.
#
#  Kontain Inc CONFIDENTIAL
#
#   This file includes unpublished proprietary source code of Kontain Inc. The
#   copyright notice above does not evidence any actual or intended publication of
#   such source code. Disclosure of this source code or any related proprietary
#   information is strictly prohibited without the express written permission of
#   Kontain Inc.
#
# Create image with kontain Faas server

#FROM alpine:3.7
FROM fedora:31

ARG DHOST

# Hack, hack hack.  We need a more general way to access the container repository inside of minikube
# or the real kube for that matter.  This stuff was just determined empirically.
ENV DOCKER_HOST=$DHOST
ENV DOCKER_CERT_PATH=/root/.minikube/certs
RUN mkdir -p /root/.minikube/certs
COPY tmp-bld/minikubecerts /root/.minikube/certs

# Go programs compiled on linux expect the GLIBC dynamic loader.
# When using "FROM alpine:3.7" we don't have the gnu loader.
# Create a symlink so the musl dynamic loader is used instead.
# See https://stackoverflow.com/questions/34729748/installed-go-binary-not-found-in-path-on-alpine-linux-docker
#RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

# Arrange for KM binaries and libraries to be in the container.
COPY tmp-bld/kontain.tar.gz /tmp/km.tgz
RUN mkdir -p /opt/kontain && tar -C /opt/kontain -xzf /tmp/km.tgz && rm /tmp/km.tgz

# Get libraries for krun
RUN dnf install -y yajl-devel libseccomp-devel

COPY kontain-faas-server /	

ENTRYPOINT [ "/kontain-faas-server" ]

## Temporary. Hard-code the functions in the file system
#RUN mkdir /kontain
#COPY test_funcs/test_func1 /kontain
#COPY test_funcs/test_func_data_with_hc.km /kontain
#COPY test_funcs/test_func_data_with_hc.json /kontain
#
## We need skopeo to pull the function containers and image tool to unpack the container into
## a bundle for krun.
## net-tools can help with network issures inside the container
#RUN dnf install -y skopeo net-tools bind-utils iputils
#COPY tmp-bld/oci-image-tool /usr/bin
#
