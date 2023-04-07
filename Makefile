# Many mickle makes a Makefile ;) 
ifeq ($(OS),Windows_NT)
SHELL := powershell.exe
.SHELLFLAGS := -NoProfile -Command
endif

VERSION := $(shell git describe --tags --always --dirty)
GOVERSION := $(shell go env GOVERSION)
BUILDTIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GITHUBURL := https://github.com/vinser/pixyne

RELEASEOPT := --release

HOSTOS := $(shell go env GOHOSTOS)
HOSTARCH := $(shell go env GOHOSTARCH)

build_cmd = \
fyne build \
$(if $(1),$(1)) \
--metadata Version=$(VERSION) \
--metadata BuildHost=$(HOSTOS)/$(HOSTARCH) \
--metadata BuildTime=$(BUILDTIME) \
--metadata GoVersion=$(GOVERSION) \
--metadata OnGitHub=$(GITHUBURL)


all: development

# Current host build development build
development:
	$(call build_cmd,)

# Current host build release build
release:
	$(call build_cmd,$(RELEASEOPT),)

.PHONY: all development release