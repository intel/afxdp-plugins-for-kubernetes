# Copyright(c) 2022 Intel Corporation.
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

FROM golang:1.20@sha256:52921e63cc544c79c111db1d8461d8ab9070992d9c636e1573176642690c14b5 as cnibuilder
COPY . /usr/src/afxdp_k8s_plugins
WORKDIR /usr/src/afxdp_k8s_plugins
RUN apt-get update && apt-get -y install --no-install-recommends libbpf-dev=1:0.3-2 \
      && make buildcni

FROM golang:1.20-alpine@sha256:87d0a3309b34e2ca732efd69fb899d3c420d3382370fd6e7e6d2cb5c930f27f9 as dpbuilder
COPY . /usr/src/afxdp_k8s_plugins
WORKDIR /usr/src/afxdp_k8s_plugins
RUN apk add --no-cache build-base~=0.5 libbsd-dev~=0.11 \
      && apk add --no-cache libbpf-dev~=0.5 --repository=https://dl-cdn.alpinelinux.org/alpine/v3.15/community \
      && make builddp

FROM amd64/alpine:3.17@sha256:e2e16842c9b54d985bf1ef9242a313f36b856181f188de21313820e177002501
RUN apk --no-cache -U add iproute2-rdma~=6.0 acl~=2.3 \
      && apk --no-cache -U add libbpf~=0.5 --repository=http://dl-cdn.alpinelinux.org/alpine/v3.15/community
COPY --from=cnibuilder /usr/src/afxdp_k8s_plugins/bin/afxdp /afxdp/afxdp
COPY --from=dpbuilder /usr/src/afxdp_k8s_plugins/bin/afxdp-dp /afxdp/afxdp-dp
COPY --from=dpbuilder /usr/src/afxdp_k8s_plugins/images/entrypoint.sh /afxdp/entrypoint.sh
ENTRYPOINT ["/afxdp/entrypoint.sh"]
