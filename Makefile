BIN_DIR = seccomp_main/

.PHONY = all clean install-dependencies

.DEFAULT_GOAL = all

CFLAGS = -Wall -Wextra -Wno-format -O2 -std=c11 -pedantic
CPPFLAGS = -I$(DDPPATH)/lib/

seccomp_main.o: $(BIN_DIR)seccomp_main.c
	$(CC) $(CFLAGS) -c $(CPPFLAGS) $< -o $@

seccomp_exec: $(BIN_DIR)seccomp_exec.c
	$(CC) $(CFLAGS) $< -o $@ -lseccomp

all: seccomp_main.o seccomp_exec
	go build -o Spielplatz ./server

clean:
	rm -f seccomp_main.o seccomp_exec Spielplatz

node_modules:
	npm install

install-dependencies: node_modules
	sudo apt-get install libseccomp-dev