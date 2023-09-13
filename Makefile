export CGO_ENABLED = 0
export GOFLAGS = -trimpath

DESTDIR :=
PREFIX := /usr/local

VERSION := $(shell grep -Po '\d\.\d\.\d' main/cmd/v.go)

COMPLETION_DIR := completion
RELEASE_DIR := release
RELEASES := linux_amd64 linux_386 linux_arm64 linux_arm windows_amd64 windows_386 android_arm64

.PHONY: all
all: zwc completion

.PHONY: zwc
zwc:
	go build -o zwc main/main.go

.PHONY: completion
completion:
	mkdir -p $(COMPLETION_DIR) && zwc completion bash > $(COMPLETION_DIR)/bash

.PHONY: test
test:
	@go test -cover
	@test/test.sh

.PHONY: clean
clean:
	-rm zwc
	-rm -r $(COMPLETION_DIR)
	-rm -r $(RELEASE_DIR)

.PHONY: install
install:
	install -Dm755 zwc $(DESTDIR)$(PREFIX)/bin/zwc
	install -Dm644 doc/zwc.1 $(DESTDIR)$(PREFIX)/share/man/man1/zwc.1
	install -Dm644 doc/zwc.5 $(DESTDIR)$(PREFIX)/share/man/man5/zwc.5
	install -Dm644 $(COMPLETION_DIR)/bash $(DESTDIR)$(PREFIX)/share/bash-completion/completions/zwc

.PHONY: uninstall
uninstall:
	rm $(DESTDIR)$(PREFIX)/bin/zwc
	rm $(DESTDIR)$(PREFIX)/share/man/man1/zwc.1
	rm $(DESTDIR)$(PREFIX)/share/man/man5/zwc.5
	rm $(DESTDIR)$(PREFIX)/share/bash-completion/completions/zwc

.PHONY: release
release: clean
	mkdir -p $(RELEASE_DIR)
	for release in ${RELEASES}; do \
		export GOOS=$${release%_*}; \
		export GOARCH=$${release#*_}; \
		export OUTFILE=$(RELEASE_DIR)/zwc_v$(VERSION)_$$release; \
		if [ $$GOOS = "windows" ]; then \
			export OUTFILE=$$OUTFILE.exe; \
		fi; \
		go build -o $$OUTFILE main/main.go; \
	done
