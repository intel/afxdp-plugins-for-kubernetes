#!/bin/bash
set -e

TEST_NET_NS=fuzznet

cleanup() {
	echo "cleanup"
	# delete test netNS, ignore potential "does not exist"
	ip netns del $TEST_NET_NS > /dev/null 2>&1 || true
}
trap cleanup EXIT

echo "installing go-fuzz"
go get -u github.com/dvyukov/go-fuzz/go-fuzz@latest github.com/dvyukov/go-fuzz/go-fuzz-build@latest

echo "building test app"
go-fuzz-build

echo "creating test netNS"
# create test netNS, ignore potential "already exists"
ip netns add $TEST_NET_NS > /dev/null 2>&1 || true
ip netns exec $TEST_NET_NS mount --bind /proc/$$/ns/net /var/run/netns/$TEST_NET_NS

echo "running tests"
go-fuzz -bin=./cni-fuzz.zip -workdir ./outputAdd -dumpcover -func FuzzAdd & \
go-fuzz -bin=./cni-fuzz.zip -workdir ./outputDel -dumpcover -func FuzzDel
