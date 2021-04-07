# Kontain Faas

Note: `make release` in your KM source directory creates `${KM_TOP}/build/kontain.tar.gz`. In the `docker build`
line below, the assumption is `kontain.tar.gz` was copied here.

- Copy ${KM_TOP}/tools/bin/kontain-gcc to /opt/kontain/bin
- Copy ${KM_TOP}/include/km_hcalls.h to /opt/kontain/include
- Copy your minikube certs into the container build area: cd YouR_BuilD_AreA; mkdir -p minkubecerts; cp $DOCKER_CERT_PATH minikubecerts

- Point to docker daemon inside minikube: `eval $(minikube docker-env)`
- Build oci-image-tool: `bash build-oci-image-tool.bash`
- Build the faas server: `bash build-server.bash`
- Build the test functions mentioned below before running the next step
- Build kontain-faas-server docker container: `docker build -t kontain-faas-server -f Dockerfile --build-arg KM_TAR=kontain.tar.gz --build-arg DHOST=$DOCKER_HOST .`
- Create a kontain kubernetes namespace if you don't have one: kubectl create namespace kontain
- Start kontain-faas-server in minikube: `kubectl create -f server.yaml`
- Get the faas pod name by running: `kubectl get pod -n kontain`, you should see output like this:
```
[paulp@work server]$ kubectl get pods
NAME                                  READY   STATUS                  RESTARTS   AGE
kontain-faas-server-d68b7b6bb-6xqbt   1/1     Running                 0          15s
kontaind-kgtwj                        0/1     Init:ImagePullBackOff   0          48d
nginx-557dcdb56b-r65mp                1/1     Running                 3          73d
[paulp@work server]$
```
- Expose kontain-fass-server to the outside world: `kubectl port-forward -n kontain kontain-faas-server-<instance> 8080:8080`.  Replace \<instance\> from the get pods output.
- Run `curl http://localhost:8080/<path>` to test the faas service.  \<path\> is the name of the function plus any arguments.  
The name is expected to be the name of an executable in the test_funcs directory.  The .km exention is added to the name by the faas server.
I've currently tested with the function name: test_func_data_with_hc
- When you are done testing, interrupt `kubectl port-forward ....` with control-C.
- If you are having problems, get the faas services logs: `kubectl logs -n kontain kontain-faas-server-d68b7b6bb-6xqbt` but put the current pod name in the command.
- To stop kontain-faas-server in minikube: `kubectl delete -f server.yaml`

## Example Function

There are example functions and containers made for them in the test_funcs directory.
To build the faas test functions:

- make -C test_funcs all
