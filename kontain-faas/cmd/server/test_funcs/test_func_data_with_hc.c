#include <stdio.h>
#include "km_hcalls.h"
#define _GNU_SOURCE         /* See feature_test_macros(7) */
#include <unistd.h>
#include <sys/syscall.h>
#include <errno.h>
#include <string.h>

/*
 * A simple test prgram that uses the HC_snapshot_getdata and HC_snapshot_putdata hypercalls
 * to pretend to be a program that will be called by the faas daemon via krun and km.
 */

int main(int argc, char* argv[])
{
   char inputbuf[2048];
   ssize_t bytesread;
   char outputbuf[2048];
   ssize_t byteswritten;

   bytesread = syscall(HC_snapshot_getdata, inputbuf, sizeof(inputbuf));
   if (bytesread < 0) {
      fprintf(stderr, "%s: HC_snapshot_getdata hypercall failed, %s\n", argv[0], strerror(errno));
      return 1;
   }

   fprintf(stdout, "%s: Got %ld bytes of input data\n", argv[0], bytesread);
   fprintf(stdout, "%s", inputbuf);
   fprintf(stdout, "=== end of data ===\n");
   fflush(stdout);

   // We don't process the input data for this test.

   // Make some output data as follows.
   // [paulp@work server]$ base64 -
   // the quick brown fox
   // dGhlIHF1aWNrIGJyb3duIGZveAo=
   // [paulp@work server]$
   snprintf(outputbuf, sizeof(outputbuf),
            "STATUSCODE: 200\n"
            "DATA: dGhlIHF1aWNrIGJyb3duIGZveAo=\n");
   byteswritten = syscall(HC_snapshot_putdata, outputbuf, strlen(outputbuf));
   if (byteswritten != strlen(outputbuf)) {
      if (byteswritten < 0) {
         fprintf(stderr, "%s: HC_snapshot_putdata hypercall failed, %s\n", argv[0], strerror(errno));
      } else {
         fprintf(stderr, "%s: HC_snapshot_putdata hypercall failed, expected return value %ld, got %ld\n", argv[0], strlen(outputbuf), byteswritten);
      }
      return 2;
   }

   return 0;
}
