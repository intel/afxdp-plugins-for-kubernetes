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

package tools

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestArrayContains(t *testing.T) {
	testCases := []struct {
		str      string
		array    []string
		expected bool
	}{
		{
			str:      "ens785f2",
			array:    []string{"ens785f2", "ens785f3", "eno1", "veth7b0b36aa@if3"},
			expected: true,
		},
		{
			str:      "cni0",
			array:    []string{"ens785f2", "ens785f3", "eno1", "veth7b0b36aa@if3"},
			expected: false,
		},
		{
			str:      "eth0",
			array:    []string{"eth0"},
			expected: true,
		},
		{
			str:      "veth0",
			array:    []string{},
			expected: false,
		},
		{
			str:      "veth0",
			array:    nil,
			expected: false,
		},
	}
	for i, tc := range testCases {
		assert.Equal(t, tc.expected, ArrayContains(tc.array, tc.str), "Should be equal: test case %d", i)
	}
}

func TestRemove(t *testing.T) {
	testCases := []struct {
		rem      string
		array    []string
		expected []string
	}{
		{
			rem:      "veth0",
			array:    []string{"ens785f2", "ens785f3", "eno1", "veth7b0b36aa@if3"},
			expected: []string{"ens785f2", "ens785f3", "eno1", "veth7b0b36aa@if3"},
		},
		{
			rem:      "eno1",
			array:    []string{"ens785f2", "ens785f3", "eno1", "veth7b0b36aa@if3"},
			expected: []string{"ens785f2", "ens785f3", "veth7b0b36aa@if3"},
		},
		{
			rem:      "cni0",
			array:    []string{"ens785f2", "cni0", "cni0", "veth7b0b36aa@if3"},
			expected: []string{"ens785f2", "cni0", "veth7b0b36aa@if3"},
		},
		{
			rem:      "",
			array:    []string{},
			expected: []string{},
		},
		{
			rem:      "",
			array:    nil,
			expected: nil,
		},
	}
	for i, tc := range testCases {
		assert.Equal(t, tc.expected, Remove(tc.array, tc.rem), "Should be equal: test case %d", i)
	}
}

func TestPathExists(t *testing.T) {
	testCases := []struct {
		path     string
		expected bool
	}{
		{
			path:     "",
			expected: false,
		},
		{
			path:     "\n",
			expected: false,
		},
		{
			path:     "#",
			expected: false,
		},
		{
			path:     "/home",
			expected: true,
		},
		{
			path:     "./../internal",
			expected: true,
		},
		{
			path:     "./tools_test.go",
			expected: true,
		},
	}
	for i, tc := range testCases {
		output, _ := PathExists(tc.path)
		assert.Equal(t, tc.expected, output, "Should be equal: test case %d", i)
	}
}

func TestContainsPrefix(t *testing.T) {
	testCases := []struct {
		str      string
		array    []string
		expected bool
	}{
		{
			str:      "",
			array:    []string{},
			expected: false,
		},
		{
			str:      "",
			array:    []string{""},
			expected: true,
		},
		{
			str:      "eno1",
			array:    []string{"eno1"},
			expected: true,
		},
		{
			str:      "veth0",
			array:    []string{"cni1", "ens", "veth"},
			expected: true,
		},
	}
	for i, tc := range testCases {
		assert.Equal(t, tc.expected, ContainsPrefix(tc.array, tc.str), "Should be equal: test case %d", i)
	}
}

func TestPrettyString(t *testing.T) {
	type TestData struct {
		Str     string
		Integer int
		Array   []string
		hidden  string
	}
	makeTestData := func(str string, integer int, array []string, hidden string) TestData {
		return TestData{str, integer, array, hidden}
	}
	testCases := []struct {
		object   TestData
		expected string
	}{
		{
			object: makeTestData(`"`, -0, []string{""}, "hello world"),
			expected: strings.Replace(
				`{
				  "Str": "\"",
				  "Integer": 0,
				  "Array": [
				    ""
				  ]
				}`, "\t", "", -1),
		},
		{
			object: makeTestData("afxdp", 30, []string{"veth1", "cni0"}, "veth7b0b36aa@if3"),
			expected: strings.Replace(
				`{
				  "Str": "afxdp",
				  "Integer": 30,
				  "Array": [
				    "veth1",
				    "cni0"
				  ]
				}`, "\t", "", -1),
		},
		{
			object: makeTestData("", 0, nil, ""),
			expected: strings.Replace(
				`{
				  "Str": "",
				  "Integer": 0,
				  "Array": null
				}`, "\t", "", -1),
		},
		{
			object: makeTestData("//", 034, []string{}, ""),
			expected: strings.Replace(
				`{
				  "Str": "//",
				  "Integer": 28,
				  "Array": []
				}`, "\t", "", -1),
		},
	}
	for i, tc := range testCases {
		output, _ := PrettyString(tc.object)
		assert.Equal(t, tc.expected, output, "Should be equal: test case %d", i)
	}
}
