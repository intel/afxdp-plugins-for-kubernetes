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

#include "log.h"
#include "_cgo_export.h"

// log function definitions, a wrapper for GoLogger
void Log_Debug(char msg[]) {
    GoLogger(msg, LOG_DEBUG);
}
void Log_Info(char msg[]) {
    GoLogger(msg, LOG_INFO);
}
void Log_Warning(char msg[]) {
    GoLogger(msg, LOG_WARN);
}
void Log_Error(char msg[]) {
    GoLogger(msg, LOG_ERROR);
}
void Log_Panic(char msg[]) {
    GoLogger(msg, LOG_PANIC);
}


// log level getters definitions
int Get_log_debug() { return LOG_DEBUG; }
int Get_log_info() { return LOG_INFO; }
int Get_log_warn() { return LOG_WARN; }
int Get_log_error() { return LOG_ERROR; }
int Get_log_panic() { return LOG_PANIC; }