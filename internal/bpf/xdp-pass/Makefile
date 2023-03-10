# TODO COPYRIGHT
LLC ?= llc
CLANG ?= clang

all: xdppass

xdppass:
	$(CLANG) -S \
	-target bpf \
	-D __BPF_TRACING__ \
	-I/usr/include/bpf \
	-Wall \
	-Wno-unused-value \
	-Wno-pointer-sign \
	-Wno-compare-distinct-pointer-types \
	-Werror \
	-O2 -emit-llvm -c -g -o xdp_pass.ll xdp_pass.c
	$(LLC) -march=bpf -filetype=obj -o xdp_pass.o xdp_pass.ll

clean:
	rm -f *.o xdp_pass.ll
