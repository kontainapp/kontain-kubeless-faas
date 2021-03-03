# Kontain Faas

Note: `make release` in your KM source directory creates `${KM_TOP}/build/kontain.tar.gz`. In the `docker build`
line below, the assumption is `kontain.tar.gz` was copied here.

- Point to docker daemon inside minikube: `eval $(minikube docker-env)`
- Build kontain-faas-server docker container: `docker build -t kontain-faas-server -f Dockerfile --build-arg KM_TAR=kontain.tar gz .`
- Start kontain-faas-server in minikube: `kubectl create -f server.yaml`
- Expose kontain-fass-server to the outside world: `kubectl port-forward -n kontain kontain-faas-server-<instance> 8080:8080`. Use `kubectl get pod -n kontain` to get pod name. `curl http://localhost:8080/<path>` will be passed to the service.
- Stop kontain-faas-server in minikube: `kubectl delete -f server.yaml`

## Example Function

- `make test_func1` to build.
