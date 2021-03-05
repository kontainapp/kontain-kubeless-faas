/*
 * usage: test1 
 */

#define _GNU_SOURCE
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <sys/syscall.h>

char output_buffer[4096];

// Hardcoded from km_hcalls.h
#define HC_snapshot_getdata 504
#define HC_snapshot_putdata 503

char outbuf[4096];
int
main(int argc, char *argv[])
{
  char *msg = getenv("FUNC_MSG");
  if (msg == NULL) {
    fprintf(stderr, "Env variable FUNC_MSG does not exist\n");
    return 1;
  }
  snprintf(outbuf, sizeof(outbuf), "STATUSCODE: 200\nDATA: %s\n", msg);
  printf("%s", outbuf);
  int msglen = strlen(outbuf);
  syscall(HC_snapshot_putdata, msg, msglen);
  return 0;
}
