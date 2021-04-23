
SERVER_NAME     := glmemo
TARGET_DIR      := bin
BUILD_NAME      := linux-amd64-${SERVER_NAME}
BUILD_VERSION   := $(shell date "+%Y%m%d.%H%M%S")
BUILD_TIME      := $(shell date "+%F %T")
COMMIT_SHA1     := $(shell git rev-parse HEAD )

all: release

install:

release:
	CGO_ENABLED=1 \
	go build -ldflags \
	"-s -w \
	-X '${SERVER_NAME}/config.Version=${BUILD_VERSION}' \
	-X '${SERVER_NAME}/config.BuildTime=${BUILD_TIME}' \
	-X '${SERVER_NAME}/config.CommitID=${COMMIT_SHA1}' \
	" -o ${TARGET_DIR}/${BUILD_NAME}

	# upx ${TARGET_DIR}/${BUILD_NAME}
	# rm -rf ${TARGET_DIR}/${BUILD_NAME}.upx

update: release
	cd bin && ./installer.sh && cd ..

update_wdz: release
	cd bin && ./installer_wdz.sh && cd ..