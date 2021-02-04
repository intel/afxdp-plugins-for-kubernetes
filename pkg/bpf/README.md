To build the static library of the wrapper c program:
$ gcc wrapper.c -lbpf -c
$ ar rs libwrapper.a wrapper.o

Then you should be able to:
$ go build
$ go run wrapper.go
