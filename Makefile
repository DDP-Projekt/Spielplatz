BIN_DIR = seccomp_main/

.PHONY = all clean install-dependencies

.DEFAULT_GOAL = all

CFLAGS = -Wall -Wextra -Wno-format -O2 -std=c11 -pedantic
CPPFLAGS = -I$(DDPPATH)/lib/
DDPVERSION = $(shell kddp version | head -n1 | cut -d ' ' -f1)
SITE_BUILD_PATH = ./site/build

INSTALL=apt-get install

unsec_main.o: $(BIN_DIR)unsec_main.c
	$(CC) $(CFLAGS) -c $(CPPFLAGS) $< -o $@

seccomp_main.o: $(BIN_DIR)seccomp_main.c
	$(CC) $(CFLAGS) -c $(CPPFLAGS) $< -o $@

seccomp_exec: $(BIN_DIR)seccomp_exec.c
	$(CC) $(CFLAGS) $< -o $@ -lseccomp

all: seccomp_main.o seccomp_exec unsec_main.o site install-dependencies
	go build -o Spielplatz -ldflags "-X main.DDPVERSION=$(DDPVERSION)" ./server

clean:
	rm -f seccomp_main.o seccomp_exec Spielplatz
	rm -rf $(SITE_BUILD_PATH)

site:
	npm --prefix site install
	npm --prefix site run build

install-dependencies:
	$(INSTALL) libseccomp-dev