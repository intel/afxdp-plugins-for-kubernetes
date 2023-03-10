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

#ifndef _WRAPPER_H_
#define _WRAPPER_H_

int Load_bpf_send_xsk_map(char *ifname);
int Load_attach_bpf_xdp_pass(char *ifname);
int Configure_busy_poll(int fd, int busy_timeout, int busy_budget);
int Clean_bpf(char *ifname);

#endif
