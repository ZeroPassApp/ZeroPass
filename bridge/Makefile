LIB_NAME ?= libzeropass
OUT_DIR ?= .
SDKROOT ?= $(shell xcrun --sdk macosx --show-sdk-path)
MACOSX_DEPLOYMENT_TARGET ?= 14.0

CGO_MIN_VERSION = -mmacosx-version-min=$(MACOSX_DEPLOYMENT_TARGET)

.PHONY: build-arm64 build-amd64 build-universal clean

build-arm64:
	CGO_ENABLED=1 GOARCH=arm64 CC=clang \
	CGO_CFLAGS="-isysroot $(SDKROOT) -arch arm64 $(CGO_MIN_VERSION)" \
	CGO_LDFLAGS="-isysroot $(SDKROOT) -arch arm64 $(CGO_MIN_VERSION)" \
	go build -buildmode=c-archive -o "$(OUT_DIR)/$(LIB_NAME)_arm64.a" .

build-amd64:
	CGO_ENABLED=1 GOARCH=amd64 CC=clang \
	CGO_CFLAGS="-isysroot $(SDKROOT) -arch x86_64 $(CGO_MIN_VERSION)" \
	CGO_LDFLAGS="-isysroot $(SDKROOT) -arch x86_64 $(CGO_MIN_VERSION)" \
	go build -buildmode=c-archive -o "$(OUT_DIR)/$(LIB_NAME)_amd64.a" .

build-universal: build-arm64 build-amd64
	lipo -create "$(OUT_DIR)/$(LIB_NAME)_arm64.a" "$(OUT_DIR)/$(LIB_NAME)_amd64.a" -output "$(OUT_DIR)/$(LIB_NAME).a"
	cp -f "$(OUT_DIR)/$(LIB_NAME)_arm64.h" "$(OUT_DIR)/$(LIB_NAME).h"

clean:
	rm -f \
		"$(OUT_DIR)/$(LIB_NAME).a" \
		"$(OUT_DIR)/$(LIB_NAME).h" \
		"$(OUT_DIR)/$(LIB_NAME)_arm64.a" \
		"$(OUT_DIR)/$(LIB_NAME)_arm64.h" \
		"$(OUT_DIR)/$(LIB_NAME)_amd64.a" \
		"$(OUT_DIR)/$(LIB_NAME)_amd64.h"
