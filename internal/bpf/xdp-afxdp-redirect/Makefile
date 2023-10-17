# Copyright(c) Red Hat Inc.
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

LLC ?= llc
CLANG ?= clang

all: afxdp_redirect

afxdp_redirect:
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
