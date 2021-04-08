#ifndef _WRAPPER_H_
#define _WRAPPER_H_

#include <stdio.h>

int load_bpf_send_xsk_map (char *ifname);
void cleanbpf (char *ifname);

// log level constants
static const int LOG_INFO = 0;
static const int LOG_WARN = 1;
static const int LOG_ERROR = 2;

// logging getter declarations
int GET_LOG_INFO();
int GET_LOG_WARN();
int GET_LOG_ERROR();

#endif

