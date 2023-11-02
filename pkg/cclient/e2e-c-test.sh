#!/bin/bash

# Build binary, .h, .a and .so files
go build -o lib_udsclient.so -buildmode=c-shared ./cclient.go
gcc -c cTest.c -o cTest.o
cp lib_udsclient.so /lib
gcc -pthread -o cTest cTest.o -L. /lib/lib_udsclient.so
# Build and push image
docker build -t localhost:5000/afxdp-c-test:latest .
docker push localhost:5000/afxdp-c-test