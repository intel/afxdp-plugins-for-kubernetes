/*
 Copyright(c) 2021 Intel Corporation.
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package main

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetConfig(t *testing.T) {

	testCases := []struct {
		name       string
		configFile string
		expConfig  *config
		expErr     error
	}{
		{
			name:       "load good config 1",
			configFile: `{"pools":[{"name":"pool1","devices":[ "dev_1", "dev_2" ]}]}`,
			expConfig:  &config{Pools: []poolConfig{{Name: "pool1", Devices: []string{"dev_1", "dev_2"}}}},
		},
		{
			name:       "load good config 2",
			configFile: `{"pools":[{"name":"pool1","devices":[ "dev_1", "dev_2" ]},{"name":"pool2","devices":["dev_3","dev_4"]}]}`,
			expConfig:  &config{Pools: []poolConfig{{Name: "pool1", Devices: []string{"dev_1", "dev_2"}}, {Name: "pool2", Devices: []string{"dev_3", "dev_4"}}}},
		},
		{
			name:       "load no config",
			configFile: `{}`,
			expConfig:  &config{Pools: []poolConfig(nil)},
		},
		{
			name:       "load bad config 1",
			configFile: ``,
			expConfig:  nil,
			expErr:     errors.New("unexpected end of JSON input"),
		},
		{
			name:       "load bad config 2",
			configFile: `{`,
			expConfig:  nil,
			expErr:     errors.New("unexpected end of JSON input"),
		},
		{
			name:       "load bad config 3",
			configFile: `{"pools"}`,
			expConfig:  nil,
			expErr:     errors.New("invalid character '}' after object key"),
		},
		{
			name:       "load bad config 4",
			configFile: `{"pools":}`,
			expConfig:  nil,
			expErr:     errors.New("invalid character '}' looking for beginning of value"),
		},
		{
			name:       "load bad config 5",
			configFile: `{"pools",}`,
			expConfig:  nil,
			expErr:     errors.New("invalid character ',' after object key"),
		},
		{
			name:       "load bad config 6",
			configFile: `{"pools":[{"name":"pool1","devices":[ "dev_1", "dev_2" ]}]`,
			expConfig:  nil,
			expErr:     errors.New("unexpected end of JSON input"),
		},
		{
			name:       "load bad config 7",
			configFile: `{"pools":[{"name":"pool1","devices":[ "dev_1", "dev_2" ]}}`,
			expConfig:  nil,
			expErr:     errors.New("invalid character '}' after array element"),
		},
		{
			name:       "load bad config 8",
			configFile: `{"pools":[{"name":"pool1","devices":[ "dev_1", "dev_2" }]}`,
			expConfig:  nil,
			expErr:     errors.New("invalid character '}' after array element"),
		},
	}
	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {

			rawConfig := []byte(tc.configFile)
			cfg, err := getConfig(rawConfig)

			if err == nil {
				assert.Equal(t, tc.expErr, err, "Error was expected")
			} else {
				require.Error(t, tc.expErr, "Unexpected error returned")
				assert.Contains(t, err.Error(), tc.expErr.Error(), "Unexpected error returned")
			}
			assert.Equal(t, tc.expConfig, cfg, "Returned unexpected config")

		})
	}
}
