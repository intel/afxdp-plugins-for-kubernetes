FROM amd64/alpine:3.14

ENV PKG_CONFIG_PATH=/usr/lib64/pkgconfig
ENV LD_LIBRARY_PATH=/usr/lib

# Environment variables for building libbpf
ENV LD_CONFIG_DIR=/etc/ld.so.conf.d
ENV LIBBPF_DIR=/libbpf
ENV LIBBPF_SRC_DIR=/libbpf/src
ENV LIBBPF_GIT_REPO=https://github.com/libbpf/libbpf.git

RUN apk -U add build-base libbsd-dev git elfutils-dev gcompat

# Clone, build, and install libbpf
RUN git clone ${LIBBPF_GIT_REPO} ${LIBBPF_DIR} \
      && make -C ${LIBBPF_SRC_DIR} > /dev/null \
      && make install -C ${LIBBPF_SRC_DIR} > /dev/null \
      && cp -r /usr/lib64/libbpf.* /usr/lib \
      && mkdir ${LD_CONFIG_DIR} \
      && echo /usr/lib64 > ${LD_CONFIG_DIR}/x86_64-linux-gnu.conf \
      && ldconfig ${LD_CONFIG_DIR} \
      && rm -rf ${LIBBPF_DIR} \
      && apk del git

# Copy AFX Device Plugin binaries
COPY ./bin/cndp /cndp/cndp
COPY ./bin/cndp-dp /cndp/cndp-dp

ENV LD_CONFIG_DIR=
ENV LIBBPF_DIR=
ENV LIBBPF_SRC_DIR=
ENV LIBBPF_GIT_REPO=

ENTRYPOINT ["./cndp/cndp-dp", "-config", "/cndp/config/config.json"]
