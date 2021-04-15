KONTAINGCC = /opt/kontain/bin/kontain-gcc
KONTAININCL = /opt/kontain/include
KONTAIN_LIB = /opt/kontain/lib

TOP := $(shell git rev-parse --show-toplevel)
C_TEMPLATE_PATH = ${TOP}/kontain-faas/function/c/
C_HEADER_PATH = -I ${KONTAININCL} -I ${C_TEMPLATE_PATH}


all:	test_func1 test_func_data_with_hc.km

test_func1: test_func1.c

test_func_data_with_hc.km:	test_func_data_with_hc.c
	${KONTAINGCC} -o test_func_data_with_hc.km -I${KONTAININCL} test_func_data_with_hc.c

test_func_km: test_func_km.c

hello.km: hello.c
	${KONTAINGCC} ${C_HEADER_PATH} -o $@ $< ${C_TEMPLATE_PATH}/km_function_template.c ${KONTAIN_LIB}/libkontain.a

clean:
	rm -f test_func1 test_func_km test_func_data_with_hc.km hello.km

images: test_func1 test_func_km test_func_data_with_hc.km
	docker build -t test_func_1 -f test_func_1.dockerfile  .
	docker build -t test_func_2 -f test_func_2.dockerfile  .
	docker build -t test_func_data_with_hc -f test_func_data_with_hc.dockerfile .
