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

FROM amd64/alpine:3.14 as bpf
RUN apk --no-cache -U add libbpf=0.4.0-r0

FROM bpf as builder
COPY . /usr/src/afxdp_k8s_plugins
WORKDIR /usr/src/afxdp_k8s_plugins
RUN apk add --no-cache go=1.16.10-r0 make=4.3-r0 libbsd-dev=0.11.3-r0 libbpf-dev=0.4.0-r0 \
      && make builddp

FROM bpf
COPY --from=builder /usr/src/afxdp_k8s_plugins/bin/afxdp /afxdp/afxdp
COPY --from=builder /usr/src/afxdp_k8s_plugins/bin/afxdp-dp /afxdp/afxdp-dp
COPY --from=builder /usr/src/afxdp_k8s_plugins/images/entrypoint.sh /afxdp/entrypoint.sh
ENTRYPOINT ["/afxdp/entrypoint.sh"]
