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


.PHONY: fmt build install run release
