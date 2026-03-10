#include "runtime/include/DDP/runtime.h"
#include <errno.h>
#include <stdio.h>

extern int ddp_ddpmain();

int main(int argc, char* argv[]) {
    ddp_init_runtime(argc, argv);
    #ifdef WINNT
    // _IOLBF doesn't work on Windows and sets full buffering instead
    // TODO: no buffering also doesn't do anything :(
    if (setvbuf(stdout, NULL, _IONBF, 0) != 0) {
        fputs("setvbuf(stdout) failed\n", stderr);
    }
    if (setvbuf(stderr, NULL, _IONBF, 0) != 0) {
        fputs("setvbuf(stderr) failed\n", stderr);
    }
    #else
    // set line buffering
    setvbuf(stdout, NULL, _IOLBF, 0);
    setvbuf(stderr, NULL, _IOLBF, 0);
    #endif
    int ret = ddp_ddpmain();
    ddp_end_runtime();
    return ret;
}