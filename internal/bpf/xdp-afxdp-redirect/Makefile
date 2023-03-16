# TODO COPYRIGHT
LLC ?= llc
CLANG ?= clang

all: xdpafxdp

xdpafxdp:
	$(CLANG) -S \
	-target bpf \
	-D __BPF_TRACING__ \
	-I/usr/include/bpf \
	-Wall \
	-Wno-unused-value \
	-Wno-pointer-sign \
	-Wno-compare-distinct-pointer-types \
	-Werror \
	-O2 -emit-llvm -c -g -o xdp_afxdp_redirect.ll xdp_afxdp_redirect.c
	$(LLC) -march=bpf -filetype=obj -o xdp_afxdp_redirect.o xdp_afxdp_redirect.ll

clean:
	rm -f *.o xdp_afxdp_redirect.ll
