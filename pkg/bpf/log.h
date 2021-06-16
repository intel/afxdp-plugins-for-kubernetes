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

#ifndef _LOG_H_
#define _LOG_H_

#include <stdio.h>

// log level constants
static const int LOG_DEBUG = 0;
static const int LOG_INFO = 1;
static const int LOG_WARN = 2;
static const int LOG_ERROR = 3;
static const int LOG_PANIC = 4;

// log level getter declarations
int Get_log_debug();
int Get_log_info();
int Get_log_warn();
int Get_log_error();
int Get_log_panic();

// log function declaration
void Log_Debug(char msg[]);
void Log_Info(char msg[]);
void Log_Warning(char msg[]);
void Log_Error(char msg[]);
void Log_Panic(char msg[]);

#endif