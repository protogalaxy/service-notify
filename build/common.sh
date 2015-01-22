#!/bin/bash

set -eu -o pipefail

PG_ROOT=$(pwd)
cd $PG_ROOT

source "${PG_ROOT}/build/golang.sh"

RUN_ENV="${RUN_ENV:-local}"

readonly LOCAL_TARGET_ROOT="${PG_ROOT}/target"

readonly DOCKER_TARGET_ROOT="/target"

if [[ "${RUN_ENV}" == "docker" ]]; then
  readonly TARGET_ROOT="${DOCKER_TARGET_ROOT}"
else
  readonly TARGET_ROOT="${LOCAL_TARGET_ROOT}"
fi

readonly TARGET_BIN="${TARGET_ROOT}/bin"
readonly TARGET_TEST="${TARGET_ROOT}/test"
readonly TARGET_TEST_COVERAGE="${TARGET_TEST}/coverage"

readonly PG_BUILD_IMAGE_NAME=protogalaxy-build

function pg::build::verify() {
  echo "+++ Verifying prerequisites ..."

  if [[ -z "$(which docker)" ]]; then
    echo "Can't find docker executable in PATH." >&2
    exit 1
  fi
}

function pg::build::build_image() {
  echo "+++ Builing docker image: ${PG_BUILD_IMAGE_NAME}"
  docker build -t "${PG_BUILD_IMAGE_NAME}" "$PG_ROOT/build/build-image"
}

function pg::build::run_build() {
  local -r project_path=$1

  echo "+++ Running build command ..."
  rm -rf cid
  docker run --cidfile=cid \
    -v "$(pwd)":"${project_path}" \
    -w "${project_path}" \
    "${PG_BUILD_IMAGE_NAME}" godep go build -o /target/bin/main

  local cid
  cid=$(cat cid)

  echo "+++ Copying built files from the build container: ${cid}"
  docker cp $cid:$DOCKER_TARGET_ROOT $PG_ROOT
  echo "+++ Removing the build container: ${cid}"
  docker rm $cid 2> /dev/null || true
  rm -rf cid
}

function pg::build::run_command() {
  local -r cmd=$1

  local -r path="${GOPATH}/${PROJECT_PATH}"

  echo "+++ Running command ..."
  rm -rf cid
  docker run --cidfile=cid \
    -e RUN_ENV=docker \
    -v "$(pwd)":"${path}" \
    -w "${path}" \
    "${PG_BUILD_IMAGE_NAME}" "$cmd"

  local cid
  cid=$(cat cid)

  echo "+++ Copying built files from the build container"
  docker cp $cid:$DOCKER_TARGET_ROOT $PG_ROOT

  echo "+++ Removing the build container"
  docker rm $cid 2> /dev/null || true
  rm -rf cid
}

function pg::build::build_release_image() {
  local -r image_name=$1

  echo "+++ Builing docker image: ${image_name}"
  docker build -t "${image_name}" "$PG_ROOT"
}
