# Kontain FAAS

Kontain FAAS is a Proof of Concept for a Kubernetes based FAAS platform that leverages the Kontain 
Unikernel/Virtual Machine technology. Like other FAAS platforms, each user function are encapsulated in 
a dedicated OCI image. When a function is called the function's OCI-image is instantiated in a Kontainer
and the function Kontainer is called. This Kontainer only lives for the time it takes to service
that single function call.

The central component of Kontain FAAS is  a kubernetes `Deployment` called`kontain-faas-server`.
Each deployed `kontain-faas-server` pod contains:

* A file system for that contains directories for the OCI-images and OCI-bundles of the user functions to be run.
* A container running the `faas-server`, which handles calls to user functions.
- A container running the `faas-downloader` which monitors for changes to Image CRD records.


# Kubless faas

Repository for shared Faas work within Kontain based on Kubeless. The real code is in submodules. Do not forget to `git submodule update --init` after cloning.

Submodules: 

- kubeless : core kubeless engine
- runtimes : runtime containers
- http-trigger : triggers involving 
- cronjob-trigger

# Overview of Vanilla Kubeless

Kubeless is Kubernetes native in the sense that all Kubeless functionality is wrapped in containers and those
containers are deployed using native Kubernetes primitives. The major container types used by kubeless are:

- `kubeless-function-controller` reacts to changes in Kubeless Function CRD's. Builds and deploys function
containers in the cluster.
- `http-trigger-controller` reacts to changes in Kubeless HTTP Trigger CRD's. Manipilates Kubernestes Ingress
resource(s) to implement Kubeless HTTP trigger.
- `cronjob-trigger-controller` reacts to changes in Kubeless CronJon Trigger CRD's. Manipulates standard Kubernetes
CronJob resource to implement Kubeless CronJob Trigger.

These three containers are run together in the Kubeless management pod(s).

When Kubeless a function is introduced to the system a function image is created by combining the user's code
with a pre-existing 'runtime' image. (Note: a signature tag on the container is used to recognize when a function
image already exists). A Kubernetes Job is used to build a function container. That job uses the following
containers  which are defined in the configuration.

- Provision Image. (default `kubeless/unzip`)
- Builder Image. (default `kubeless/function-image-builder`) used for a step of a function container build.
This container uses `skopeo` to combine a runtime template container with the user's function. The result is
pushed to the container registry, again using `skopeo`.

Kubeless is deployed with the command `kubectl create -f kubeless.yaml`. The `kubeless.yaml` file is built by running `make -C kubeless kubeless.yaml`.

## Build Targets

- `make -C kubeless function-controller` builds kubeless function controller image 
(`kubeless-function-controller`).
- `make -C cronjob-trigger cronjob-controller-image` builds the kubeless CronJob trigger controller image (`cronjob-trigger-controller`).
- `make -C http-trigger http-controller-image` builds the kubeless HTTP trigger controller image (`http-trigger-controller`).
- `make -C kubeless function-image-builder` builds the container `kubeless-function-image-builder:latest`.
- `make -C runtimes/stable/python build3.8` builds the runtime container for python 3.8 (`kubeless/python:3.8`).
