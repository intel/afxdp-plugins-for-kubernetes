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

#include <bpf/xsk.h>	   // for xsk_setup_xdp_prog, bpf_set_link_xdp_fd
#include <linux/if_link.h> // for XDP_FLAGS_UPDATE_IF_NOEXIST
#include <net/if.h>	   // for if_nametoindex

#include "bpfWrapper.h"
#include "log.h"

#define SO_PREFER_BUSY_POLL 69
#define SO_BUSY_POLL_BUDGET 70

int Load_bpf_send_xsk_map(char *ifname) {
	int fd = -1;
	int if_index, ret;

	Log_Info("Load_bpf_send_xsk_map(): disovering if_index for interface %s", ifname);

	if_index = if_nametoindex(ifname);
	if (!if_index) {
		Log_Error("Load_bpf_send_xsk_map(): if_index not valid: %s", ifname);
		return -1;
	} else {
		Log_Info("Load_bpf_send_xsk_map(): if_index for interface %s is %d", ifname,
			 if_index);
	}

	Log_Info("Load_bpf_send_xsk_map(): starting setup of xdp program on "
		 "interface %s (%d)",
		 ifname, if_index);

	ret = xsk_setup_xdp_prog(if_index, &fd);
	if (ret) {
		Log_Error("Load_bpf_send_xsk_map(): setup of xdp program failed, "
			  "returned: %d",
			  ret);
		return -1;
	}

	if (fd > 0) {
		Log_Info("Load_bpf_send_xsk_map(): loaded xdp program on interface %s "
			 "(%d), file descriptor %d",
			 ifname, if_index, fd);
		return fd;
	}

	return -1;
}

int Configure_busy_poll(int fd, int busy_timeout, int busy_budget) {
	int sock_opt = 1;
	int ret;

	Log_Info("Configure_busy_poll(): setting SO_PREFER_BUSY_POLL on file descriptor %d", fd);

	ret = setsockopt(fd, SOL_SOCKET, SO_PREFER_BUSY_POLL, (void *)&sock_opt, sizeof(sock_opt));
	if (ret < 0) {
		Log_Error("Configure_busy_poll(): failed to set SO_PREFER_BUSY_POLL on file "
			  "descriptor %d, returned: %d",
			  fd, ret);
		return 1;
	}

	Log_Info("Configure_busy_poll(): setting SO_BUSY_POLL to %d on file descriptor %d",
		 busy_timeout, fd);

	sock_opt = busy_timeout;
	ret = setsockopt(fd, SOL_SOCKET, SO_BUSY_POLL, (void *)&sock_opt, sizeof(sock_opt));
	if (ret < 0) {
		Log_Error("Configure_busy_poll(): failed to set SO_BUSY_POLL on file descriptor "
			  "%d, returned: %d",
			  fd, ret);
		goto err_timeout;
	}

	Log_Info("Configure_busy_poll(): setting SO_BUSY_POLL_BUDGET to %d on file descriptor %d",
		 busy_budget, fd);

	sock_opt = busy_budget;
	ret = setsockopt(fd, SOL_SOCKET, SO_BUSY_POLL_BUDGET, (void *)&sock_opt, sizeof(sock_opt));
	if (ret < 0) {
		Log_Error("Configure_busy_poll(): failed to set SO_BUSY_POLL_BUDGET on file "
			  "descriptor %d, returned: %d",
			  fd, ret);
	} else {
		Log_Info("Configure_busy_poll(): busy polling budget on file descriptor %d set to "
			 "%d",
			 fd, busy_budget);
		return 0;
	}

	Log_Warning("Configure_busy_poll(): setsockopt failure, attempting to restore xsk to "
		    "default state");

	Log_Warning("Configure_busy_poll(): unsetting SO_BUSY_POLL on file descriptor %d", fd);

	sock_opt = 0;
	ret = setsockopt(fd, SOL_SOCKET, SO_BUSY_POLL, (void *)&sock_opt, sizeof(sock_opt));
	if (ret < 0) {
		Log_Error("Configure_busy_poll(): failed to unset SO_BUSY_POLL on file descriptor "
			  "%d, returned: %d",
			  fd, ret);
		return 1;
	}

err_timeout:
	Log_Warning("Configure_busy_poll(): unsetting SO_PREFER_BUSY_POLL on file descriptor %d",
		    fd);
	sock_opt = 0;
	ret = setsockopt(fd, SOL_SOCKET, SO_PREFER_BUSY_POLL, (void *)&sock_opt, sizeof(sock_opt));
	if (ret < 0) {
		Log_Error("Configure_busy_poll(): failed to unset SO_PREFER_BUSY_POLL on file "
			  "descriptor %d, returned: %d",
			  fd, ret);
		return 1;
	}
}

int Clean_bpf(char *ifname) {
	int if_index, ret;
	int fd = -1;

	Log_Info("Clean_bpf(): disovering if_index for interface %s", ifname);

	if_index = if_nametoindex(ifname);
	if (!if_index) {
		Log_Error("Clean_bpf(): if_index not valid: %s", ifname);
		return 1;
	} else {
		Log_Info("Clean_bpf(): if_index for interface %s is %d", ifname, if_index);
	}

	Log_Info("Clean_bpf(): starting removal of xdp program on interface %s (%d)", ifname,
		 if_index);

	ret = bpf_set_link_xdp_fd(if_index, fd, XDP_FLAGS_UPDATE_IF_NOEXIST);
	if (ret) {
		Log_Error("Clean_bpf(): Removal of xdp program failed, returned: ", ret);
		return 1;
	}

	Log_Info("Clean_bpf(): removed xdp program from interface %s (%d)", ifname, if_index);
	return 0;
}
