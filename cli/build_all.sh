#!/bin/sh

BIN_DIR=binaries
BINARY_NAME=a8cli
STATIC_BINARY=0
TARGET_OS="linux windows darwin"

if [[ "$BIN_DIR" ]]; then
  rm -rf ${BIN_DIR}
  mkdir ${BIN_DIR}
fi

echo "Building binaries..."

for os in ${TARGET_OS}; do
  case "$os" in
    "linux")
      echo "1. linux"
      env GOOS=${os} GOARCH=amd64 CGO_ENABLED=${STATIC_BINARY} go build -o ${BIN_DIR}/${BINARY_NAME}
      cd ${BIN_DIR}
      tar -cf "${BINARY_NAME}-${os}.tgz" ${BINARY_NAME}
      cd ..
      ;;

    "windows")
      echo "2. windows"
      env GOOS=${os} GOARCH=amd64 CGO_ENABLED=${STATIC_BINARY} go build -o "${BIN_DIR}/${BINARY_NAME}.exe"
      cd ${BIN_DIR}
      zip -q "${BINARY_NAME}-${os}" "${BINARY_NAME}.exe"
      rm "${BINARY_NAME}.exe"
      cd ..
      continue
      ;;

    "darwin")
      echo "3. mac"
      env GOOS=${os} GOARCH=amd64 CGO_ENABLED=${STATIC_BINARY} go build -o ${BIN_DIR}/${BINARY_NAME}
      cd ${BIN_DIR}
      tar -cf "${BINARY_NAME}-mac.tgz" ${BINARY_NAME}
      cd ..
      ;;
  esac
  rm ${BIN_DIR}/${BINARY_NAME}
done

echo "done."
