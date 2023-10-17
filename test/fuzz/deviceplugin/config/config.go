/*
 * Copyright(c) 2022 Intel Corporation.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package deviceplugin

import (
	"io/ioutil"
	"os"

	dp "github.com/intel/afxdp-plugins-for-kubernetes/internal/deviceplugin"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/dpcnisyncerserver"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/host"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/networking"
)

const (
	tempDirectory  = "config/" //temp directory is created upon fuzz.sh execution
	udsDirFileMode = os.FileMode(0700)
)

var firstRun bool = true

/*
Fuzz sends fuzzed data into the GetConfig function
The input data is considered:
  - uninteresting if is caught by an existing error
  - interesting if it does not result in an error, input priority increases for subsequent fuzzing
  - discard if it will not unmarshall, so we don't just end up testing the json.Unmarshall function
*/
func Fuzz(data []byte) int {
	if firstRun {
		firstRun = false
		if err := os.MkdirAll(tempDirectory, udsDirFileMode); err != nil {
			panic(1)
		}
	}

	tmpfile, err := ioutil.TempFile(tempDirectory, "config_")
	if err != nil {
		os.Remove(tmpfile.Name())
		panic(1)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write(data); err != nil {
		os.Remove(tmpfile.Name())
		panic(1)
	}
	if err := tmpfile.Close(); err != nil {
		os.Remove(tmpfile.Name())
		panic(1)
	}

	//START THE SYNCER SERVER TODO CHECK BPF MAP
	dpCniSyncerServer, err := dpcnisyncerserver.NewSyncerServer()
	if err != nil {
		panic(1)
	}

	_, err = dp.GetPoolConfigs(tmpfile.Name(), networking.NewHandler(), host.NewHandler(), dpCniSyncerServer)
	if err != nil {
		return 0
	}

	return 1

}
