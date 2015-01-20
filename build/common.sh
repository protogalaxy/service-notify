#!/usr/bin/bash

set -eu -o pipefail

PG_ROOT=$(pwd)
cd $PG_ROOT

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
  trap "rm -rf cid" EXIT

  echo "+++ Running build command ..."
  docker run --cidfile=cid \
    -v "$(pwd)":"${project_path}" \
    -w "${project_path}" \
    "${PG_BUILD_IMAGE_NAME}" godep go build -o /build/main

  readonly CID=$(cat cid)
  echo "+++ Copying built files from the build container: ${CID}"
  docker cp $CID:/build/main .
  echo "+++ Removing the build container: ${CID}"
  docker rm $CID > /dev/null
}

function pg::build::build_release_image() {
  local -r image_name=$1

  echo "+++ Builing docker image: ${image_name}"
  docker build -t "${image_name}" "$PG_ROOT"
}
