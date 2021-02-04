/* SPDX-License-Identifier: BSD-3-Clause
 * Copyright(c) 2020 Intel Corporation.
 */
#include <bpf/bpf.h>
#include <bpf/libbpf.h>        // for bpf_get_link_xdp_id, bpf_set_link_xdp_fd
#include <bpf/xsk.h>
#include <bsd/string.h>           // for strlcpy
#include <getopt.h>               // for no_argument, getopt_long, option
#include <linux/if_link.h>        // for XDP_FLAGS_UPDATE_IF_NOEXIST
#include <net/if.h>               // for IFNAMSIZ
#include <poll.h>                 // for pollfd, poll, POLLIN
#include <pthread.h>
#include <signal.h>            // for signal, SIGUSR1, SIGINT
#include <stdio.h>             // for printf, fprintf, NULL, stderr, size_t
#include <stdlib.h>            // for free, exit, realloc, EXIT_SUCCESS
#include <string.h>            // for memcpy, strcmp, strlen
#include <sys/socket.h>        // for accept, bind, listen, socket, AF_UNIX
#include <sys/stat.h>
#include <sys/types.h>
#include <sys/un.h>

#include "wrapper.h"       // for golang wrapper of bpf code

int load_bpf_send_xsk_map (char *ifname)
{
    int fd = -1;
    int if_index, ret;

    if_index = if_nametoindex(ifname);
        if (!if_index)
           printf("if_index not valid");
    printf("ifname=%s ifindex=%d\n", ifname, if_index);

    ret = xsk_setup_xdp_prog(if_index, &fd);
    if (ret)
       printf("Setup of xdp program failed ret=%d\n", ret);

    if(fd > 0){
        printf("xsks_map_fd=%d\n", fd);
        printf("loaded xdp program\n");
    }

    return fd;
}
