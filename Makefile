.PHONY: all build compress test clean

BINARY_NAME=gclone
BUILD_FLAGS=-trimpath -buildvcs=false
LDFLAGS=-s -w -buildid=
UNAME_S := $(shell uname -s)

all: build

build:
	CGO_ENABLED=0 go build $(BUILD_FLAGS) -ldflags="$(LDFLAGS)" -o bin/$(BINARY_NAME) .

compress: build
	@if [ "$(UNAME_S)" = "Darwin" ]; then \
		echo "Warning: macOS (Mach-O) does not support UPX compression (macOS dyld terminates UPX-packed binaries with SIGKILL). Skipping UPX on macOS."; \
	elif command -v upx >/dev/null 2>&1; then \
		echo "Compressing bin/$(BINARY_NAME) with UPX..."; \
		upx --best --lzma bin/$(BINARY_NAME); \
	else \
		echo "UPX not installed, skipping compression"; \
	fi

test:
	go test -v ./...

clean:
	rm -rf bin/ dist/
