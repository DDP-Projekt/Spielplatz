BIN_DIR = seccomp_main/

.PHONY = all clean install-dependencies

.DEFAULT_GOAL = all

CFLAGS = -Wall -Wextra -Wno-format -O2 -std=c11 -pedantic
CPPFLAGS = -I$(DDPPATH)/lib/
DDPVERSION = $(shell kddp version | head -n1 | cut -d ' ' -f1)

INSTALL=apt-get install

seccomp_main.o: $(BIN_DIR)seccomp_main.c
	$(CC) $(CFLAGS) -c $(CPPFLAGS) $< -o $@

seccomp_exec: $(BIN_DIR)seccomp_exec.c
	$(CC) $(CFLAGS) $< -o $@ -lseccomp

all: seccomp_main.o seccomp_exec node_modules install-dependencies
	go build -o Spielplatz -ldflags "-X main.DDPVERSION=$(DDPVERSION)" ./server

clean:
	rm -f seccomp_main.o seccomp_exec Spielplatz
	rm -rf ./node_modules

node_modules:
	npm install

install-dependencies: node_modules
	$(INSTALL) libseccomp-dev