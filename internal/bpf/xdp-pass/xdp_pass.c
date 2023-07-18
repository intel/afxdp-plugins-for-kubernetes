/*
 * Copyright(c) Red Hat Inc.
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
// clang-format off
#include <linux/types.h>
#include <bpf/bpf_helpers.h>
#include <linux/bpf.h>
// clang-format on
SEC("xdp")
int xdp_prog_pass(struct xdp_md *ctx) { return XDP_PASS; }

char _license[] SEC("license") = "Dual BSD";
