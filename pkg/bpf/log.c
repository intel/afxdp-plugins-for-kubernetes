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

// 256 is the maximum number of characters for vsnprintf
#define LOG_SIZE 256

void Log_Debug(char *fmt, ...) {
	char msg[LOG_SIZE];
	va_list args;

	va_start(args, fmt);
	vsnprintf(msg, sizeof(msg), fmt, args);
	va_end(args);

	Debugf(msg);
}

void Log_Info(char *fmt, ...) {
	char msg[LOG_SIZE];
	va_list args;

	va_start(args, fmt);
	vsnprintf(msg, sizeof(msg), fmt, args);
	va_end(args);

	Infof(msg);
}

void Log_Warning(char *fmt, ...) {
	char msg[LOG_SIZE];
	va_list args;

	va_start(args, fmt);
	vsnprintf(msg, sizeof(msg), fmt, args);
	va_end(args);

	Warningf(msg);
}

void Log_Error(char *fmt, ...) {
	char msg[LOG_SIZE];
	va_list args;

	va_start(args, fmt);
	vsnprintf(msg, sizeof(msg), fmt, args);
	va_end(args);

	Errorf(msg);
}

void Log_Panic(char *fmt, ...) {
	char msg[LOG_SIZE];
	va_list args;

	va_start(args, fmt);
	vsnprintf(msg, sizeof(msg), fmt, args);
	va_end(args);

	Panicf(msg);
}
