/*
 * Copyright(c) 2021 Intel Corporation.
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

#include "log.h"

#include "_cgo_export.h"

// 256 is the maximum number of characters for vsnprintf
#define LOG_SIZE 256

void log_fn(log_level_t level, char *fmt, ...) {
	char msg[LOG_SIZE];
	va_list args;
	va_start(args, fmt);
	vsnprintf(msg, sizeof(msg), fmt, args);
	va_end(args);

	switch (level) {
	case LOG_LEVEL_ERR:
		Errorf(msg);
		break;
	case LOG_LEVEL_INFO:
		Infof(msg);
		break;
	case LOG_LEVEL_DEBUG:
		Debugf(msg);
		break;
	case LOG_LEVEL_PANIC:
		Panicf(msg);
		break;
	case LOG_LEVEL_WARNING:
		Warningf(msg);
		break;
	}
}
