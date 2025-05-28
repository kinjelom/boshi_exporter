#!/bin/bash

PROGRAM_NAME="boshi_exporter"
VERSION="0.0.5"

DIST_DIR=".dist"

declare -a os_array=("linux") # "darwin" "windows")
declare -a arch_array=("amd64") # "arm64")

mkdir -p "$DIST_DIR"

for OS in "${os_array[@]}"; do
    for ARCH in "${arch_array[@]}"; do
        FULL_NAME="${PROGRAM_NAME}_${VERSION}.${OS}-${ARCH}"
        mkdir -p "${DIST_DIR}/${FULL_NAME}"
        DIST_PATH="${DIST_DIR}/${FULL_NAME}/${PROGRAM_NAME}"
        echo "build $DIST_PATH"
        if GOOS=$OS GOARCH=$ARCH go build -o "$DIST_PATH" -ldflags="-X 'main.ProgramVersion=${VERSION}'" >> "${DIST_PATH}.log"; then
            sha256sum "$DIST_PATH"  | awk '{print $1}' > "${DIST_PATH}.sum"
        fi
        tar -czvf "${DIST_DIR}/${FULL_NAME}.tar.tgz" -C "$DIST_DIR" "$FULL_NAME"
    done
done


