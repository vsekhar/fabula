SRCS = $(wildcard images/*.mmd)
OBJS = $(SRCS:.mmd=.png)

# From: https://stackoverflow.com/questions/21708839/problems-setting-path-in-makefile/26936855#26936855
PATH := $(PATH):$(HOME)/node_modules/.bin
SHELL := env PATH=$(PATH) /bin/bash

MERMAID=mmdc
MERMAIDCONFIG=images/mermaidconfig.json

%.png: %.mmd $(MERMAIDCONFIG)
	$(MERMAID) --configFile $(MERMAIDCONFIG) --input $< --output $@

default: $(OBJS)

all: clean default

.PHONY: clean default

clean:
	rm -rf $(OBJS)
