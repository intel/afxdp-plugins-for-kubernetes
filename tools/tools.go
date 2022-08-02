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
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

/*
ArrayContains returns true if str is an element of array.
*/
func ArrayContains(array []string, str string) bool {
	for _, s := range array {
		if s == str {
			return true
		}
	}
	return false
}

/*
ContainsPrefix returns true if str is prefixed with any element of array.
*/
func ContainsPrefix(array []string, str string) bool {
	for _, s := range array {
		if strings.HasPrefix(str, s) {
			return true
		}
	}
	return false
}

/*
PathExists returns true if path is exists, false if non-existent.
*/
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

/*
Remove returns array without the element rem if it is present.
*/
func Remove(array []string, rem string) []string {
	for i, elm := range array {
		if elm == rem {
			return append(array[:i], array[i+1:]...)
		}
	}
	return array
}

/*
PrettyString formats v as a string for logging purposes.
*/
func PrettyString(v interface{}) (string, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return fmt.Sprint(string(b)), nil
}
