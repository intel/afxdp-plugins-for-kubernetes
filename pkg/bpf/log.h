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

#ifndef _LOG_H_
#define _LOG_H_

#include <stdarg.h>
#include <stdio.h>

#define foreach_log_level                                                                          \
	_(0, PANIC, panic)                                                                         \
	_(1, ERR, err)                                                                             \
	_(2, WARNING, warn)                                                                        \
	_(3, INFO, info)                                                                           \
	_(4, DEBUG, debug)

typedef enum {
#define _(n, uc, lc) LOG_LEVEL_##uc = n,
	foreach_log_level
#undef _
} log_level_t;

void log_fn(log_level_t level, char *fmt, ...);

#define Log_Panic(fmt, ...) log_fn(LOG_LEVEL_PANIC, fmt, __VA_ARGS__)
#define Log_Error(fmt, ...) log_fn(LOG_LEVEL_ERR, fmt, __VA_ARGS__)
#define Log_Warning(fmt, ...) log_fn(LOG_LEVEL_WARNING, fmt, __VA_ARGS__)
#define Log_Info(fmt, ...) log_fn(LOG_LEVEL_INFO, fmt, __VA_ARGS__)
#define Log_Debug(fmt, ...) log_fn(LOG_LEVEL_DEBUG, fmt, __VA_ARGS__)

#endif