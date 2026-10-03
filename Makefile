.PHONY: all build clean
export PACKETRUSHER_VERSION PACKETRUSHER_REVISION PACKETRUSHER_BUILD_TIME
all: build
build:
	./scripts/build.sh packetrusher
clean:
	rm -f ./packetrusher
