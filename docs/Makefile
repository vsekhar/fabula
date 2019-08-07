SRCS = $(wildcard images/*.mmd)
OBJS = $(SRCS:.mmd=.png)

MERMAID=mmdc

%.png: %.mmd
	echo $(PWD); echo $(PATH); which mmdc; $(MERMAID) -i $< -o $@

all: $(OBJS)
