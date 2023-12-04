# Copyright(c) 2022 Intel Corporation.
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

FROM golang:1.20@sha256:efe38cb419e2b2012f66d1782d2efe2fd8884c71d9f342581e1697ba9047b5f8 as cnibuilder
COPY . /usr/src/afxdp_k8s_plugins
WORKDIR /usr/src/afxdp_k8s_plugins
RUN apt-get update \
&& apt-get -y install --no-install-recommends libxdp-dev=1.3.1-1 \
&& apt-get -y install -o APT::Keep-Downloaded-Packages=false --no-install-recommends clang=1:14.0-55.7~deb12u1 \
&& apt-get -y install -o APT::Keep-Downloaded-Packages=false --no-install-recommends llvm=1:14.0-55.7~deb12u1 \
&& apt-get -y install -o APT::Keep-Downloaded-Packages=false --no-install-recommends gcc-multilib=4:12.2.0-3 \
&& make buildcni

FROM golang:1.20-alpine@sha256:ebceb16dc094769b6e2a393d51e0417c19084ba20eb8967fb3f7675c32b45774 as dpbuilder
COPY . /usr/src/afxdp_k8s_plugins
WORKDIR /usr/src/afxdp_k8s_plugins
RUN apk add --no-cache build-base~=0.5-r3 \
&& apk add --no-cache libbsd-dev~=0.11.7 \
&& apk add --no-cache libxdp-dev~=1.2.10-r0 \
&& apk add --no-cache libbpf-dev~=1.0.1-r0 \
&& apk add --no-cache llvm15~=15.0.7-r0 \
&& apk add --no-cache clang15~=15.0.7-r0 \
&& make builddp

FROM amd64/alpine:3.18@sha256:25fad2a32ad1f6f510e528448ae1ec69a28ef81916a004d3629874104f8a7f70
RUN apk --no-cache -U add iproute2-rdma~=6.3.0-r0 acl~=2.3 \
      && apk add --no-cache xdp-tools~=1.2.10-r0
COPY --from=cnibuilder /usr/src/afxdp_k8s_plugins/bin/afxdp /afxdp/afxdp
COPY --from=dpbuilder /usr/src/afxdp_k8s_plugins/bin/afxdp-dp /afxdp/afxdp-dp
COPY --from=dpbuilder /usr/src/afxdp_k8s_plugins/images/entrypoint.sh /afxdp/entrypoint.sh
COPY --from=dpbuilder /usr/src/afxdp_k8s_plugins/internal/bpf/xdp-pass/xdp_pass.o /afxdp/xdp_pass.o
COPY --from=dpbuilder /usr/src/afxdp_k8s_plugins/internal/bpf/xdp-afxdp-redirect/xdp_afxdp_redirect.o /afxdp/xdp_afxdp_redirect.o
ENTRYPOINT ["/afxdp/entrypoint.sh"]
