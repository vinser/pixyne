# Many mickle makes a Makefile ;) 
ifeq ($(OS),Windows_NT)
SHELL := powershell.exe
.SHELLFLAGS := -NoProfile -Command
endif

GOVERSION := $(shell go env GOVERSION)
BUILDTIME := $(shell date -u +"%Y-%m-%d %H:%M:%S")

HOSTOS := $(shell go env GOHOSTOS)
HOSTARCH := $(shell go env GOHOSTARCH)

ifeq ($(HOSTOS),linux)
# semantic_ver = $(shell sh/app-semver.sh)
$(shell sh/app-semver.sh)

xbuild_cmd = \
$(HOME)/go/bin/fyne-cross \
$(1) \
-arch=$(2) \
-pull \
-metadata BuildForOS="$(1)/$(2)" \
-metadata BuildTime="$(BUILDTIME)" \
-metadata GoVersion="$(GOVERSION)"
endif

## -----------------------------------------------------------------
##  
## Usage: make <target>
##  
## where target is one of:
##  
.NOTPARALLEL:
.PHONY: help
help:
ifeq ($(HOSTOS),linux)
	@sed -n '/@sed/!s/:.*##//p;s/^## //p' $(MAKEFILE_LIST)
else
	@echo For cross-builds app use linux OS host 
endif

# Cross-build on linux OS host
ifeq ($(HOSTOS),linux)
include MakefileX
else
xbuild xdarwin xlinux xwindows macsdk xdarwin_amd64 xdarwin_arm64 xlinux_amd64 xlinux_arm64 xlinux_arm xlinux_386 xwindows_amd64 xwindows_386:
	@echo For cross-builds app use linux OS host 
endif