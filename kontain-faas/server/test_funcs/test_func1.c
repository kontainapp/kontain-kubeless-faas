/*
 * usage: test1 <request file> <response file>
 */

#include <errno.h>
#include <stdio.h>

int
main(int argc, char *argv[])
{
  char *request_path = argv[1];
  char *response_path = argv[2];

  if (argc != 3) {
    fprintf(stderr, "usage: %s <request_path> <response_path>\n", argv[0]);
    return 1;
  }
  printf("request: %s response %s\n", request_path, response_path);
  FILE* resp_f = fopen(response_path, "w");
  if (resp_f == NULL) {
    fprintf(stderr, "fopen(%s, \"w\") failed - %d\n", errno);
    return 1;
  }
  fprintf(resp_f, "STATUSCODE: 200\n");
  // SGVsbG8gV29ybGQK is "Hello World\n" base64 encoded.
  fprintf(resp_f, "DATA: SGVsbG8gV29ybGQK\n");
  fclose(resp_f);
  return 0;
}
