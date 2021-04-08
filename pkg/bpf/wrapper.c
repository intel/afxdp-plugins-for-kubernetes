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
#include <bpf/bpf.h>
#include <bpf/libbpf.h> // for bpf_get_link_xdp_id, bpf_set_link_xdp_fd
#include <bpf/xsk.h>
#include <linux/if_link.h> // for XDP_FLAGS_UPDATE_IF_NOEXIST
#include <net/if.h>        // for IFNAMSIZ
#include <stdio.h>         // for printf, fprintf, NULL, stderr, size_t
#include <stdlib.h>        // for free, exit, realloc, EXIT_SUCCESS
#include <string.h>        // for memcpy, strcmp, strlen

#include "_cgo_export.h"
#include "wrapper.h"

#define LOG_SIZE 256

// log level getters definition
int GET_LOG_INFO() { return LOG_INFO; }
int GET_LOG_WARN() { return LOG_WARN; }
int GET_LOG_ERROR() { return LOG_ERROR; }

char buf[LOG_SIZE];

int load_bpf_send_xsk_map(char *ifname) {
  int fd = -1;
  int if_index, ret;

  if_index = if_nametoindex(ifname);
  if (!if_index) {
    snprintf(buf, sizeof buf, "%s%s",
             "func load_bpf_send_xsk_map: if_index not valid: ", ifname);
    cLogger(buf, LOG_ERROR);
  } else {
    snprintf(buf, sizeof buf, "%s%s%s%d",
             "func load_bpf_send_xsk_map: disovering if_index for interface  ",
             ifname, ", If_index for interface is: ", if_index);
    cLogger(buf, LOG_INFO);
  }

  cLogger("Starting setup of XDP program", LOG_INFO);
  ret = xsk_setup_xdp_prog(if_index, &fd);
  if (ret) {
    snprintf(buf, sizeof buf, "%s%d",
             "func load_bpf_send_xsk_map: if_index not valid: ", if_index);
    cLogger(buf, LOG_ERROR);

  } else {
    snprintf(buf, sizeof buf, "%s%d",
             "func xsk_setup_xdp_prog: setup of xdp program failed ret= ", ret);
    cLogger(buf, LOG_ERROR);
  }

  if (fd > 0) {
    snprintf(buf, sizeof buf, "%s%s%s%d%s%d",
             "func load_bpf_send_xsk_map:loaded XDP program on interface",
             ifname, "file descriptor :", fd, ",if_index: ", if_index);
    cLogger(buf, LOG_INFO);
  }

  return fd;
}

void cleanbpf(char *ifname) {
  int if_index, ret;
  int fd = -1;

  if_index = if_nametoindex(ifname);
  if (!if_index) {
    snprintf(buf, sizeof buf, "%s%d%s%s", "func cleanbpf: if_index ", if_index,
             " not valid for interface: ", ifname);
    cLogger(buf, LOG_ERROR);
  } else {
    snprintf(buf, sizeof buf, "%s%s%s%d",
             "func load_bpf_send_xsk_map: disovering if_index for interface  ",
             ifname, ", if_index is: ", if_index);
    cLogger(buf, LOG_INFO);
  }
  cLogger("Starting removal of XDP program", LOG_INFO);

  ret = bpf_set_link_xdp_fd(if_index, fd, XDP_FLAGS_UPDATE_IF_NOEXIST);
  if (ret) {
    snprintf(buf, sizeof buf, "%s%s",
             "func cleanbpf: Removal of xdp program failed on interface ",
             ifname);
    cLogger(buf, LOG_ERROR);

  } else {
    snprintf(buf, sizeof buf, "%s%s%s%d",
             "func cleanbpf: Unloaded bpf program from interface", ifname,
             " ,ret=", ret);
    cLogger(buf, LOG_INFO);
  }
}
