package cni

import (
	"encoding/json"
	"fmt"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/intel/cndp_device_plugin/cmd/cni"
	"k8s.io/apimachinery/pkg/util/uuid"
)

const (
	fuzzNs        = "fuzznet"
	interesting   = 1
	uninteresting = 0
	discard       = -1
)

/*
FuzzAdd sends fuzzed data into the cni CmdAdd function
The input data is considered:
 - uninteresting if is caught by an exesting error
 - interesting if it does not result in an error, input priority increases for subsequent fuzzing
 - discard if it will not unmarshall, so we don't just end up testing the json.Unmarshall function
*/
func FuzzAdd(data []byte) int {
	tmp := &cni.NetConfig{}
	if err := json.Unmarshal(data, tmp); err != nil {
		return discard
	}

	fakeID := uuid.NewUUID()
	a := &skel.CmdArgs{
		StdinData:   data, // data prepared by go-fuzz
		Netns:       "/var/run/netns/" + fuzzNs,
		ContainerID: string(fakeID),
		IfName:      fmt.Sprintf("eth%v", int(fakeID[7])),
	}

	if err := cni.CmdAdd(a); err != nil {
		return uninteresting
	}
	return interesting
}

/*
FuzzDel sends fuzzed data into the cni CmdDel function
The input data is considered:
 - uninteresting if is caught by an exesting error
 - interesting if it does not result in an error, input priority increases for subsequent fuzzing
 - discard if it will not unmarshall, so we don't just end up testing the json.Unmarshall function
*/
func FuzzDel(data []byte) int {
	tmp := &cni.NetConfig{}
	if err := json.Unmarshal(data, tmp); err != nil {
		return discard
	}

	fakeID := uuid.NewUUID()
	a := &skel.CmdArgs{
		StdinData:   data, // data prepared by go-fuzz
		Netns:       "/var/run/netns/" + fuzzNs,
		ContainerID: string(fakeID),
		IfName:      fmt.Sprintf("eth%v", int(fakeID[7])),
	}

	if err := cni.CmdDel(a); err != nil {
		return uninteresting
	}
	return interesting
}
