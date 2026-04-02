LIB_NAME=libzeropass

.PHONY: build-arm64 build-amd64 build-universal clean

build-arm64:
	CGO_ENABLED=1 GOARCH=arm64 go build -buildmode=c-archive -o $(LIB_NAME)_arm64.a .

build-amd64:
	CGO_ENABLED=1 GOARCH=amd64 go build -buildmode=c-archive -o $(LIB_NAME)_amd64.a .

build-universal: build-arm64 build-amd64
	lipo -create $(LIB_NAME)_arm64.a $(LIB_NAME)_amd64.a -output $(LIB_NAME).a

clean:
	rm -f $(LIB_NAME).a $(LIB_NAME).h $(LIB_NAME)_arm64.a $(LIB_NAME)_arm64.h $(LIB_NAME)_amd64.a $(LIB_NAME)_amd64.h
