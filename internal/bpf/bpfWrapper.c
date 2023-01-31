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

#include <bpf/xsk.h> // for xsk_setup_xdp_prog, bpf_set_link_xdp_fd
#include <net/if.h>  // for if_nametoindex

#include "bpfWrapper.h"
#include "log.h"

#define SO_PREFER_BUSY_POLL 69
#define SO_BUSY_POLL_BUDGET 70
#define EBUSY_CODE_WARNING -16
#define XDP_FLAGS_UPDATE_IF_NOEXIST (1U << 0)

int Load_bpf_send_xsk_map(char *ifname) {

	int fd = -1;
	int if_index, err;

	Log_Info("%s: disovering if_index for interface %s", __FUNCTION__, ifname);

	if_index = if_nametoindex(ifname);
	if (!if_index) {
		Log_Error("%s: if_index not valid: %s", __FUNCTION__, ifname);
		return -1;
	} else {
		Log_Info("%s: if_index for interface %s is %d", __FUNCTION__, ifname, if_index);
	}

	Log_Info("%s: starting setup of xdp program on "
		 "interface %s (%d)",
		 __FUNCTION__, ifname, if_index);

	err = xsk_setup_xdp_prog(if_index, &fd);
	if (err) {
		Log_Error("%s: setup of xdp program failed, "
			  "returned: %d",
			  __FUNCTION__, err);
		return -1;
	}

	if (fd > 0) {
		Log_Info("%s: loaded xdp program on interface %s "
			 "(%d), file descriptor %d",
			 __FUNCTION__, ifname, if_index, fd);
		return fd;
	}

	return -1;
}

int Configure_busy_poll(int fd, int busy_timeout, int busy_budget) {

	int sock_opt = 1;
	int err;

	Log_Info("%s: setting SO_PREFER_BUSY_POLL on file descriptor %d", __FUNCTION__, fd);

	err = setsockopt(fd, SOL_SOCKET, SO_PREFER_BUSY_POLL, (void *)&sock_opt, sizeof(sock_opt));
	if (err < 0) {
		Log_Error("%s: failed to set SO_PREFER_BUSY_POLL on file "
			  "descriptor %d, returned: %d",
			  __FUNCTION__, fd, err);
		return 1;
	}

	Log_Info("%s: setting SO_BUSY_POLL to %d on file descriptor %d", __FUNCTION__, busy_timeout,
		 fd);

	sock_opt = busy_timeout;
	err = setsockopt(fd, SOL_SOCKET, SO_BUSY_POLL, (void *)&sock_opt, sizeof(sock_opt));
	if (err < 0) {
		Log_Error("%s: failed to set SO_BUSY_POLL on file descriptor "
			  "%d, returned: %d",
			  __FUNCTION__, fd, err);
		goto err_timeout;
	}

	Log_Info("%s: setting SO_BUSY_POLL_BUDGET to %d on file descriptor %d", __FUNCTION__,
		 busy_budget, fd);

	sock_opt = busy_budget;
	err = setsockopt(fd, SOL_SOCKET, SO_BUSY_POLL_BUDGET, (void *)&sock_opt, sizeof(sock_opt));
	if (err < 0) {
		Log_Error("%s: failed to set SO_BUSY_POLL_BUDGET on file "
			  "descriptor %d, returned: %d",
			  __FUNCTION__, fd, err);
	} else {
		Log_Info("%s: busy polling budget on file descriptor %d set to "
			 "%d",
			 __FUNCTION__, fd, busy_budget);
		return 0;
	}

	Log_Warning("%s: setsockopt failure, attempting to restore xsk to default state",
		    __FUNCTION__);

	Log_Warning("%s: unsetting SO_BUSY_POLL on file descriptor %d", __FUNCTION__, fd);

	sock_opt = 0;
	err = setsockopt(fd, SOL_SOCKET, SO_BUSY_POLL, (void *)&sock_opt, sizeof(sock_opt));
	if (err < 0) {
		Log_Error("%s: failed to unset SO_BUSY_POLL on file descriptor "
			  "%d, returned: %d",
			  __FUNCTION__, fd, err);
		return 1;
	}

err_timeout:
	Log_Warning("%s: unsetting SO_PREFER_BUSY_POLL on file descriptor %d", __FUNCTION__, fd);
	sock_opt = 0;
	err = setsockopt(fd, SOL_SOCKET, SO_PREFER_BUSY_POLL, (void *)&sock_opt, sizeof(sock_opt));
	if (err < 0) {
		Log_Error("%s: failed to unset SO_PREFER_BUSY_POLL on file "
			  "descriptor %d, returned: %d",
			  __FUNCTION__, fd, err);
		return 1;
	}
	return 0;
}

int Clean_bpf(char *ifname) {
	int if_index, err;
	int fd = -1;

	Log_Info("%s: disovering if_index for interface %s", __FUNCTION__, ifname);

	if_index = if_nametoindex(ifname);
	if (!if_index) {
		Log_Error("%s: if_index not valid: %s", __FUNCTION__, ifname);
		return 1;
	} else {
		Log_Info("%s: if_index for interface %s is %d", __FUNCTION__, ifname, if_index);
	}

	Log_Info("%s: starting removal of xdp program on interface %s (%d)", __FUNCTION__, ifname,
		 if_index);

	err = bpf_set_link_xdp_fd(if_index, fd, XDP_FLAGS_UPDATE_IF_NOEXIST);
	if (err) {
		if (err == EBUSY_CODE_WARNING) {
			// unloading of XDP program found to return EBUSY error of -16 on certain
			// host libbpf versions. doesn't break functionality and this problem is
			// being investigated.
			Log_Warning("%s: Removal of xdp program is reporting error code: (%d)",
				    __FUNCTION__, err);
		} else {
			Log_Error("%s: Removal of xdp program failed, returned: (%d)", __FUNCTION__,
				  err);
			return 1;
		}
	}

	Log_Info("%s: removed xdp program from interface %s (%d)", __FUNCTION__, ifname, if_index);
	return 0;
}
