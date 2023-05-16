# Makefile
# Easy testing and building
# By J. Stuart McMurray
# Created 20230516
# Last Modified 20230516

BIN=mqtxtar

.PHONY: all $(BIN) check clean

all: check $(BIN)

$(BIN):
	go build -trimpath -ldflags "-w -s" -o $(BIN)

check:
	go test
	go vet
	staticcheck

clean:
	rm -f $(BIN)
