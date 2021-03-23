
# certificates that the faas server will need to access the function container repository
# This is minikube specific.  Not sure what to do for plain old kubeless
minikubecerts:
	@if [ "$(DOCKER_CERT_PATH)" = "" ] ;\
	then \
		echo "DOCKER_CERT_PATH must be in the environment" ; \
		echo 'Run "eval $$(minikube docker-env)" to have it defined' ; \
		exit 1; \
	fi
	mkdir -p minikubecerts
	cp $(DOCKER_CERT_PATH)/* minikubecerts

# Build the faas server
kontain-faas-server:	kontain-api-handler.go	kontain-faas.go  kontain-function-root.go
	rm -f go.mod go.sum
	go mod init kontain-faas
	GOOS=linux GOARCH=amd64 go build -o kontain-faas-server .

# Build the container image unpack tool for the faas server to use
oci-image-tool:
	# Verify that GOPATH env var is set
	@if [ "$(GOPATH)" == "" ]; \
	then \
		echo "GOPATH must be in the environment"; \
		exit 1; \
	fi
	rm -f go.mod go.sum
	go get -d github.com/opencontainers/image-tools/cmd/oci-image-tool
	make -C $(GOPATH)/src/github.com/opencontainers/image-tools  tool
	cp $(GOPATH)/src/github.com/opencontainers/image-tools/oci-image-tool .

faas-server-container:	oci-image-tool minikubecerts kontain-faas-server
	@if [ "$(DOCKER_HOST)" == "" ]; \
	then \
		echo "DOCKER_HOST must be in the environment"; \
		echo 'Run "eval $$(minikube docker-env)" to have it defined'; \
		exit 1; \
	fi
	docker build  -t  kontain-faas-server  -f  Dockerfile  --build-arg  KM_TAR=kontain.tar.gz  --build-arg DHOST=$DOCKER_HOST .
