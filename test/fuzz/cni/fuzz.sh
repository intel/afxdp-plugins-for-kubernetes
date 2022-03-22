#!/bin/bash

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
