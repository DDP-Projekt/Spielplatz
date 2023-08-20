#include "runtime/include/runtime.h"
#include <seccomp.h>
#include <errno.h>

void install_seccomp_filter() {
    // install seccomp filter that only allows some syscalls
    scmp_filter_ctx ctx;
    ctx = seccomp_init(SCMP_ACT_ERRNO(EPERM));

    seccomp_rule_add(ctx, SCMP_ACT_ALLOW, SCMP_SYS(exit), 0);
    seccomp_rule_add(ctx, SCMP_ACT_ALLOW, SCMP_SYS(exit_group), 0);
    seccomp_rule_add(ctx, SCMP_ACT_ALLOW, SCMP_SYS(brk), 0);
    seccomp_rule_add(ctx, SCMP_ACT_ALLOW, SCMP_SYS(mmap), 0);
    seccomp_rule_add(ctx, SCMP_ACT_ALLOW, SCMP_SYS(munmap), 0);
    seccomp_rule_add(ctx, SCMP_ACT_ALLOW, SCMP_SYS(mremap), 0);
    seccomp_rule_add(ctx, SCMP_ACT_ALLOW, SCMP_SYS(read), 0);
    seccomp_rule_add(ctx, SCMP_ACT_ALLOW, SCMP_SYS(write), 0);
    seccomp_rule_add(ctx, SCMP_ACT_ALLOW, SCMP_SYS(clock_gettime), 0);
    seccomp_rule_add(ctx, SCMP_ACT_ALLOW, SCMP_SYS(clock_nanosleep), 0);
    seccomp_rule_add(ctx, SCMP_ACT_ALLOW, SCMP_SYS(nanosleep), 0);

    seccomp_load(ctx);
    seccomp_release(ctx);
}

extern int ddp_ddpmain();

int main(int argc, char* argv[]) {
    ddp_init_runtime(argc, argv);
    install_seccomp_filter();
    int ret = ddp_ddpmain();
    ddp_end_runtime();
    return ret;
}