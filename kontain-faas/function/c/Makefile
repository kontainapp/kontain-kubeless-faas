
km_hello_func.km: km_hello_func.c
	/opt/kontain/bin/kontain-gcc -o km_hello_func.km km_hello_func.c

clean:
	rm -f km_hello_func.km

docker-build: km_hello_func.km
	docker build -t docker.io/muthatkontain/cfunc:latest .
