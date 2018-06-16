GO ?= go
GOPATH := $(CURDIR)
GOSRCPATH := ./src
GOBINPATH := ./bin
GOOUTPUTBIN := $(GOBINPATH)/DutyRoster
SHELL := /bin/bash

export GOPATH
export GOSRCPATH
export
.PHONY : clean

all: build

clean:
	rm -rf $(GOBINPATH)/*

build:
	$(GO) build -o $(GOOUTPUTBIN) $(GOSRCPATH)
	echo -e "\n\t**** RESULT : $$? : Build completed!!! ****\n\t**** Binary is at $$PWD/bin ****"
