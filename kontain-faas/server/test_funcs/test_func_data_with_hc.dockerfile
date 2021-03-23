FROM alpine
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

COPY test_func_data_with_hc.km /usr/bin/test_func_data_with_hc.km

ENTRYPOINT [ "/usr/bin/test_func_data_with_hc.km" ]
