/*
 copyright(c) 2021 Intel Corporation.
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

#include "log.h"
#include "_cgo_export.h"

// log function definitions, a wrapper for GoLogger
void Log(char msg[], int lvl) {
  GoLogger(msg, lvl);
}

// log level getters definitions
int Get_log_info() { return LOG_INFO; }
int Get_log_warn() { return LOG_WARN; }
int Get_log_error() { return LOG_ERROR; }
