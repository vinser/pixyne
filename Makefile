# Many mickle makes a Makefile ;) 
ifeq ($(OS),Windows_NT)
SHELL := powershell.exe
.SHELLFLAGS := -NoProfile -Command
endif

GITTAGVERSION := $(shell git describe --tags --always --dirty)
GOVERSION := $(shell go env GOVERSION)
BUILDTIME := $(shell date -u +"%Y-%m-%d %H:%M:%S")

HOSTOS := $(shell go env GOHOSTOS)
HOSTARCH := $(shell go env GOHOSTARCH)

build_cmd = \
fyne-cross  \
$(if $(1),$(1)) \
$(if $(2),$(2)) \
--app-build 1 \
--app-id com.github.vinser.pixyne \
--app-version 1.0.0 \
--icon appIcon.png \
--name Pixyne \
--metadata GitTagVersion="$(GITTAGVERSION)" \
--metadata BuildHost="$(HOSTOS)/$(HOSTARCH)" \
--metadata BuildTime="$(BUILDTIME)" \
--metadata GoVersion="$(GOVERSION)" \
--metadata OnGitHub="https://github.com/vinser/pixyne"

# Current host build
all: host
	
host:
	$(call build_cmd, $(HOSTOS))

darwin:
	$(call build_cmd, darwin) -macosx-sdk-path="C:\MyDev\macOS\XC12.5\SDKs\MacOSX11.3.sdk"

linux:
	$(call build_cmd, linux) -arch=amd64
	$(call build_cmd, linux) -arch=arm64

windows:
	$(call build_cmd, windows)

xbuild: darwin linux windows

.PHONY: all host darwin linux windows xbuild