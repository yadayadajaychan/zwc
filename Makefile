export CGO_ENABLED = 0
export GOFLAGS = -trimpath

DESTDIR := /usr/local

VERSION := 0.1.1

RELEASE_DIR := release
RELEASES := linux_amd64 linux_386 linux_arm64 linux_arm windows_amd64 windows_386 android_arm64

.PHONY: build
build:
	go build -o zwc main/main.go

.PHONY: test
test:
	@go test -cover
	@test/test.sh

.PHONY: clean
clean:
	-rm zwc
	-rm -r $(RELEASE_DIR)

.PHONY: install
install:
	install -Dm755 zwc $(DESTDIR)/bin/zwc
	install -Dm644 doc/zwc.1 $(DESTDIR)/share/man/man1/zwc.1
	install -Dm644 doc/zwc.5 $(DESTDIR)/share/man/man5/zwc.5
	install -Dm644 <(./zwc completion bash) $(DESTDIR)/share/bash-completion/completions/zwc

.PHONY: uninstall
uninstall:
	rm $(DESTDIR)/bin/zwc
	rm $(DESTDIR)/share/man/man1/zwc.1
	rm $(DESTDIR)/share/man/man5/zwc.5
	rm $(DESTDIR)/share/bash-completion/completions/zwc

.PHONY: release
release:
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
