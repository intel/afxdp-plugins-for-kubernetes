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

package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIntVersion(t *testing.T) {
	testCases := []struct {
		name      string
		version   string
		expResult int64
		expError  error
	}{
		{
			name:      "first_test_5.4.0-89-generic",
			version:   "5.4.0-89-generic",
			expResult: 500040000,
			expError:  nil,
		},
		{
			name:      "second_test_5.4.0-89",
			version:   "5.4.0-89",
			expResult: 500040000,
			expError:  nil,
		},
		{
			name:      "third_test_5.4.0-generic",
			version:   "5.4.0-generic",
			expResult: 500040000,
			expError:  nil,
		},
		{
			name:      "fourth_test_5.4.0",
			version:   "5.4.0",
			expResult: 500040000,
			expError:  nil,
		},
		{
			name:      "fifth_test_5.4.0--generic",
			version:   "5.4.0--generic",
			expResult: 500040000,
			expError:  nil,
		},
		{
			name:      "sixth_test_5.4.0-43-.generic",
			version:   "5.4.0-43-.generic",
			expResult: 500040000,
			expError:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			actualReturn, err := intVersion(tc.version)

			assert.Equal(t, tc.expResult, actualReturn, "Returned value does not match expected value")

			if err != nil {
				require.Error(t, tc.expError, err, "Error was expected")
				assert.Contains(t, err.Error(), tc.expError.Error(), "Unexpected error returned")
			}
		})
	}

}
