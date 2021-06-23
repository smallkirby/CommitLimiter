GOCMD=go
GOTEST=$(GOCMD) test -v
GOBUILD=$(GOCMD) build
BINARY_NAME=smgithub

run:
	$(GOCMD) run main.go

build:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME)

release:
	$(GOCMD) mod tidy -v
	$(GOBUILD) -ldflags "-s -w" -o $(BINARY_NAME)

fmt: 
	find . -type f -name "*.go" | xargs -i $(GOCMD) fmt {}

install: Makefile
	$(MAKE) release
	sudo cp ./smgithub /usr/bin/${BINARY_NAME}
	sudo bash -c 'echo -e "SHELL=/bin/sh\n0 * * * *  root /usr/bin/smgithub\n" > /etc/cron.d/smgithub'
	echo "\033[0;32mTYPE below command to init configuration.\033[0m"
	echo "\033[0;32m  smgithub --init --username <YOUR USERNAME> --limit <NUM> \033[0m"


.PHONY: fmt build install run release
