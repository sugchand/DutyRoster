GO ?= go
GOPATH := $(CURDIR)
GOSRCPATH := ./src/DutyRoster
GOBINPATH := ./bin
GOOUTPUTBIN := $(GOBINPATH)/DutyRoster
SHELL := /bin/bash
DEP := $(shell command -v dep  2> /dev/null)

export GOPATH
export GOSRCPATH
export
.PHONY : clean

all: build

clean:
	rm -rf $(GOBINPATH)/*

build:
ifndef DEP
$(error "dep is not available please install go dep package manager")
endif
	-@(cd $(GOSRCPATH);$(DEP) status 2> /dev/null)
	@echo -e "\n\tSet 'GOPATH' to '$(GOPATH)'"
	@echo -e "\tRun 'dep ensure' in $(GOSRCPATH) to install missing third party packages\n"
	$(GO) build -o $(GOOUTPUTBIN) $(GOSRCPATH)
	@echo -e "\n\t**** RESULT : $$? : Build completed!!! ****\n\t**** Binary is at $$PWD/bin ****"
