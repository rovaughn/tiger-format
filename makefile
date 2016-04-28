#!/bin/sh -ue

.PHONY: all

all: tiger.tgr.png tiger.tgr.xz tiger.tgr.bz2

tiger: main.go
	go build

tiger.tgr: tiger tiger.png
	./tiger encode <tiger.png >tiger.tgr

tiger.tgr.png: tiger tiger.tgr
	./tiger decode <tiger.tgr >tiger.tgr.png

tiger.tgr.xz: tiger.tgr
	xz -9kf tiger.tgr

tiger.tgr.bz2: tiger.tgr
	bzip2 -9kf tiger.tgr

clean:
	rm -f tiger tiger.tgr tiger.tgr.png tiger.tgr.xz tiger.tgr.bz2

