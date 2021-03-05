
test_func1: test_func1.c

test_func_km: test_func_km.c

clean:
	rm test_func1 test_func_km

images: test_func_km
	docker build -t test_func_1 -f test_func_1.dockerfile  .
	docker build -t test_func_2 -f test_func_2.dockerfile  .
