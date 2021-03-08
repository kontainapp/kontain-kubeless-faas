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

ARG KM_TAR

# Go programs compiled on linux expect the GLIBC dynamic loader.
# When using "FROM alpine:3.7" we don't have the gnu loader.
# Create a symlink so the musl dynamic loader is used instead.
# See https://stackoverflow.com/questions/34729748/installed-go-binary-not-found-in-path-on-alpine-linux-docker
#RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

# Arrange for KM binaries and libraries to be in the container.
COPY $KM_TAR /tmp/km.tgz
RUN mkdir -p /opt/kontain && tar -C /opt/kontain -xzf /tmp/km.tgz && rm /tmp/km.tgz

COPY kontain-faas-server /	

ENTRYPOINT [ "/kontain-faas-server" ]

# Temporary. Hard-code the functions in the file system
RUN mkdir /kontain
COPY test_funcs/test_func1 /kontain
COPY test_funcs/test_func_data_with_hc.km /kontain

# Get libraries for krun
RUN dnf install -y yajl-devel libseccomp-devel

# Get the kontain programs into the container
COPY faaskrun.sh /opt/kontain/bin
