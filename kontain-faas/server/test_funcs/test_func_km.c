/*
 * usage: test1 
 */

#include <stdio.h>
#include <stdlib.h>
#include <sys/syscall.h>

char output_buffer[4096];

// Hardcoded from km_hcalls.h
#define HC_snapshot_getdata 504
#define HC_snapshot_putdata 503

int
main(int argc, char *argv[])
{
  char *msg = getenv("FUNC_MSG");
  printf("%s\n", msg);
  int msglen = strlen(msg);
  syscall(HC_snapshot_putdata, msg, msglen);
  return 0;
}
