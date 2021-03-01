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

FROM alpine:3.7

# Go programs compiled on linux expect the GLIBC dynamic loader.
# Create a symlink so the musl synamic loader is used instead.
# See https://stackoverflow.com/questions/34729748/installed-go-binary-not-found-in-path-on-alpine-linux-docker
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

COPY kontain-faas-server /	

ENTRYPOINT [ "/kontain-faas-server" ]

# Temporary. Hard-code a function in the file system
RUN mkdir /kontain
COPY test_funcs/test_func1 /kontain
