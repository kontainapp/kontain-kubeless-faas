
#define _GNU_SOURCE
#include <string.h>
#include <unistd.h>
#include <sys/syscall.h>

int main(int argc, char *argv[])
{
    char inmsg[4092];
    syscall(504, inmsg, sizeof(inmsg));
    char *outmsg = "{\"Status\": 200, \"Data\": \"aGVsbG8gZnJvbSBD\"}";
    syscall(503, outmsg, strlen(outmsg));
}
