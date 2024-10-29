#include "runtime/include/runtime.h"
#include <errno.h>
#include <seccomp.h>
#include <stdio.h>

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

int main(int argc, char *argv[]) {
  install_seccomp_filter();
  ddp_init_runtime(argc, argv);
  setvbuf(stdout, NULL, _IOLBF, 0);
  setvbuf(stderr, NULL, _IOLBF, 0);
  int ret = ddp_ddpmain();
  ddp_end_runtime();
  return ret;
}