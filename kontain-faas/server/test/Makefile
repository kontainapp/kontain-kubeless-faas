# Copyright © 2021 Kontain Inc. All rights reserved.

# Kontain Inc CONFIDENTIAL

#  This file includes unpublished proprietary source code of Kontain Inc. The
#  copyright notice above does not evidence any actual or intended publication of
#  such source code. Disclosure of this source code or any related proprietary
#  information is strictly prohibited without the express written permission of
#  Kontain Inc.

#  Makefile for sequential testing of a server
# 'make help' for help

PORT ?= 8090
ENTRY ?= localhost
END ?= healthz
REQ ?= 10
RD ?= 10

build: ## builds go executable
	go build

start-test: test ## runs the test for localhost:8090/healthz making 10 requests with 10ms delays inbetween (default)
	./test -port=${PORT} -entry=${ENTRY} -end=${END} -req=${REQ} -rd=${RD}
	