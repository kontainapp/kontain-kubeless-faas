# Copyright © 2021 Kontain Inc. All rights reserved.
#
# Kontain Inc CONFIDENTIAL
#
#  This file includes unpublished proprietary source code of Kontain Inc. The
#  copyright notice above does not evidence any actual or intended publication of
#  such source code. Disclosure of this source code or any related proprietary
#  information is strictly prohibited without the express written permission of
#  Kontain Inc

KM_TAR=${HOME}/workspace/km/build/kontain.tar.gz

all: km-tar minikubecerts faas-server-container
#	rm ./kontain.tar.gz

km-tar: 
	[ -d tmp-bld ] || mkdir tmp-bld
	cp ${KM_TAR} tmp-bld/kontain.tar.gz

# certificates that the faas server will need to access the function container repository
# This is minikube specific.  Not sure what to do for plain old kubeless
minikubecerts:
	@if [ "$(DOCKER_CERT_PATH)" = "" ] ;\
	then \
		echo "DOCKER_CERT_PATH must be in the environment" ; \
		echo 'Run "eval $$(minikube docker-env)" to have it defined' ; \
		exit 1; \
	fi
	mkdir -p tmp-bld/minikubecerts
	cp $(DOCKER_CERT_PATH)/* tmp-bld/minikubecerts

# Build the faas server
kontain-faas-server: kontain-api-handler.go kontain-faas.go
	GOOS=linux GOARCH=amd64 go build -o kontain-faas-server .

faas-server-container:	test-func-images minikubecerts kontain-faas-server
	@if [ "$(DOCKER_HOST)" == "" ]; \
	then \
		echo "DOCKER_HOST must be in the environment"; \
		echo 'Run "eval $$(minikube docker-env)" to have it defined'; \
		exit 1; \
	fi
	docker build  -t  kontain-faas-server  -f  Dockerfile  --build-arg DHOST=${DOCKER_HOST} .

test-func-images:
	make -C test_funcs images

test:
	kubectl create -f server.yaml
	sleep 5
	bats tests/test.bats
	kubectl delete -f server.yaml

clean:
	rm -f kontain-faas-server
