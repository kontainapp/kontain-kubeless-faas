# Kontain Faas

- Point to docker daemon inside minikube: `eval $(minikube docker-env)`
- Build kontain-faas-server docker container: `docker build -t kontain-faas-server .`
- Start kontain-faas-server in minikube: `kubectl create -f server.yaml`
- Expose kontain-fass-server to the outside world: `kubectl port-forward -n kontain kontain-faas-server-<instance> 8080:8080`. Use `kubectl get pod -n kontain` to get pod name. `curl http://localhost:8080/<path>` will be passed to the service.
- Stop kontain-faas-server in minikube: `kubectl delete -f server.yaml`