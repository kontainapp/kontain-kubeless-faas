
FROM alpine
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
# echo Func 2 Rules | base64
ENV FUNC_MSG "RnVuYyAyIFJ1bGVzCg=="

COPY test_func_km /test_func_km

ENTRYPOINT [ "/test_func_km" ]
