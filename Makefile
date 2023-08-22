BIN = seccomp_main.o
BIN_DIR = seccomp_main/

.PHONY = all clean install-dependencies

.DEFAULT_GOAL = all

CFLAGS = -c -Wall -Wextra -Wno-format -O2 -std=c11 -pedantic
CPPFLAGS = -I$(DDPPATH)/lib/

$(BIN): $(BIN_DIR)seccomp_main.c
	$(CC) $(CFLAGS) $(CPPFLAGS) $< -o $@

all: $(BIN)
	go build -o Spielplatz ./server

clean:
	rm -f $(BIN) Spielplatz

install-dependencies:
	sudo apt-get install libseccomp-dev
	npm install