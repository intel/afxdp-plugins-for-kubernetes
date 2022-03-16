FROM amd64/alpine:3.14 as bpf
RUN apk --no-cache -U add ethtool=5.12-r0 libbpf=0.4.0-r0

FROM bpf as builder
COPY . /usr/src/afxdp_k8s_plugins
WORKDIR /usr/src/afxdp_k8s_plugins
RUN apk add --no-cache go=1.16.10-r0 make=4.3-r0 libbsd-dev=0.11.3-r0 libbpf-dev=0.4.0-r0 \
      && make build

FROM bpf
COPY --from=builder /usr/src/afxdp_k8s_plugins/bin/afxdp /afxdp/afxdp
COPY --from=builder /usr/src/afxdp_k8s_plugins/bin/afxdp-dp /afxdp/afxdp-dp
COPY --from=builder /usr/src/afxdp_k8s_plugins/images/entrypoint.sh /afxdp/entrypoint.sh
ENTRYPOINT ["/afxdp/entrypoint.sh"]
