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

#include <bpf/bpf.h>
#include <bpf/libbpf.h> // for bpf_get_link_xdp_id, bpf_set_link_xdp_fd
#include <bpf/xsk.h>
#include <linux/if_link.h> // for XDP_FLAGS_UPDATE_IF_NOEXIST
#include <net/if.h>        // for IFNAMSIZ
#include <stdio.h>         // for printf, fprintf, NULL, stderr, size_t
#include <stdlib.h>        // for free, exit, realloc, EXIT_SUCCESS
#include <string.h>        // for memcpy, strcmp, strlen

#include "log.h"
#include "wrapper.h"

#define LOG_SIZE 256

char log_buf[LOG_SIZE];

int Load_bpf_send_xsk_map(char *ifname) {
  int fd = -1;
  int if_index, ret;

  if_index = if_nametoindex(ifname);
  if (!if_index) {
    snprintf(log_buf, sizeof log_buf, "%s%s",
             "func Load_bpf_send_xsk_map: if_index not valid: ", ifname);
    Log_Error(log_buf);
  } else {
    snprintf(log_buf, sizeof log_buf, "%s%s%s%d",
             "func Load_bpf_send_xsk_map: disovering if_index for interface  ",
             ifname, ", if_index for interface is: ", if_index);
    Log_Info(log_buf);
  }
  Log_Info("starting setup of XDP program");

  ret = xsk_setup_xdp_prog(if_index, &fd);
  if (ret) {
    snprintf(log_buf, sizeof log_buf, "%s%d",
             "func xsk_setup_xdp_prog: setup of xdp program failed ret: ", ret);
    Log_Error(log_buf);
  }

  if (fd > 0) {
    snprintf(log_buf, sizeof log_buf, "%s%s%s%d%s%d",
             "func Load_bpf_send_xsk_map:loaded XDP program on interface: ",
             ifname, "file descriptor: ", fd, ",if_index: ", if_index);
    Log_Info(log_buf);
  }

  return fd;
}

void Clean_bpf(char *ifname) {
  int if_index, ret;
  int fd = -1;

  if_index = if_nametoindex(ifname);
  if (!if_index) {
    snprintf(log_buf, sizeof log_buf, "%s%d%s%s", "func Clean_bpf: if_index ",
             if_index, " not valid for interface: ", ifname);
    Log_Error(log_buf);
  } else {
    snprintf(log_buf, sizeof log_buf, "%s%s%s%d",
             "func Load_bpf_send_xsk_map: disovering if_index for interface  ",
             ifname, ", if_index is: ", if_index);
    Log_Info(log_buf);
  }
  Log_Info("starting removal of XDP program");

  ret = bpf_set_link_xdp_fd(if_index, fd, XDP_FLAGS_UPDATE_IF_NOEXIST);
  if (ret) {
    snprintf(log_buf, sizeof log_buf, "%s%s",
             "func Clean_bpf: Removal of xdp program failed on interface ",
             ifname);
    Log_Error(log_buf);

  } else {
    snprintf(log_buf, sizeof log_buf, "%s%s%s%d",
             "func Clean_bpf: Unloaded bpf program from interface", ifname,
             " ,ret=", ret);
    Log_Info(log_buf);
  }
}
