#include "lib_udsclient.h"
#include <stdio.h>

int main() {
	printf("C Library: Client Version: %s \n", GetUdsClientVersion());
	printf("C Library: Server Version: %s \n", GetUdsServerVersion());
	printf("C Library: Xsk Map FD request: %d \n", RequestXskMapFd("enp94s0f0"));
	CleanUpConnection();
}