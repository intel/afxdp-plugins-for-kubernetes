/* SPDX-License-Identifier: BSD-3-Clause
 * Copyright(c) 2021 Intel Corporation.
 */
#include <bpf/bpf.h>
#include <bpf/libbpf.h> // for bpf_get_link_xdp_id, bpf_set_link_xdp_fd
#include <bpf/xsk.h>
#include <linux/if_link.h> // for XDP_FLAGS_UPDATE_IF_NOEXIST
#include <net/if.h>        // for IFNAMSIZ
#include <stdio.h>         // for printf, fprintf, NULL, stderr, size_t
#include <stdlib.h>        // for free, exit, realloc, EXIT_SUCCESS
#include <string.h>        // for memcpy, strcmp, strlen

#include "wrapper.h" // for golang wrapper of bpf code

int load_bpf_send_xsk_map(char *ifname) {
  int fd = -1;
  int if_index, ret;

  if_index = if_nametoindex(ifname);
  if (if_index)
    printf("if_index not valid");
  printf("ifname=%s ifindex=%d\n", ifname, if_index);

  ret = xsk_setup_xdp_prog(if_index, &fd);
  if (ret)
    printf("Setup of xdp program failed ret=%d\n", ret);

  if (fd > 0) {
    printf("xsks_map_fd=%d\n", fd);
    printf("loaded xdp program\n");
  }

  return fd;
}

void cleanbpf(char *ifname) {
  int if_index, ret;
  int fd = -1;

  // verify IF_INDEX from Interface passed to func
  if_index = if_nametoindex(ifname);
  if (if_index)
    printf("if_index not valid");
  printf("ifname=%s ifindex=%d\n", ifname, if_index);

  printf("Unloading bpf program from interface %s\n", ifname);
  ret = bpf_set_link_xdp_fd(if_index, fd, XDP_FLAGS_UPDATE_IF_NOEXIST);
  if (ret)
    printf("Removal of xdp program failed\n");
  printf("Unloaded bpf program from interface %s\n", ifname);
}
