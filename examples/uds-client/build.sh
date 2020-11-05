set -e

go build -o uds-client ./main.go

docker build \
	--build-arg http_proxy=${http_proxy} \
	--build-arg HTTP_PROXY=${HTTP_PROXY} \
	--build-arg https_proxy=${https_proxy} \
	--build-arg HTTPS_PROXY=${HTTPS_PROXY} \
	--build-arg no_proxy=${no_proxy} \
	--build-arg NO_PROXY=${NO_PROXY} \
	-t alpine-uds-test -f Dockerfile .